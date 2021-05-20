package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/host-diagnostics/hostdiag"
	"github.com/hashicorp/host-diagnostics/products"
	"github.com/hashicorp/host-diagnostics/seeker"
	"github.com/hashicorp/host-diagnostics/util"
)

// TODO: NewDryAgent() to simplify all the 'if d.Dryrun's ??

func NewAgent(logger hclog.Logger) Agent {
	a := Agent{
		l:       logger,
		results: make(map[string]interface{}),
	}
	a.start()
	return a
}

// Agent holds our set of seekers to be executed and their results.
type Agent struct {
	l       hclog.Logger
	seekers []*seeker.Seeker
	results map[string]interface{}
	tmpDir  string

	Manifest
}

// Manifest holds the metadata for a diagnostics run for rendering later.
type Manifest struct {
	Start      time.Time
	End        time.Time
	Duration   string
	NumErrors  int
	NumSeekers int

	Flags
}

// Flags stores our CLI inputs.
type Flags struct {
	OS          string
	Dryrun      bool
	Consul      bool
	Nomad       bool
	TFE         bool
	Vault       bool
	AllProducts bool
	Includes    []string
	Outfile     string
}

type CSVFlag struct {
	Values *[]string
}

func (s CSVFlag) String() string {
	if s.Values == nil {
		return ""
	}
	return strings.Join(*s.Values, ",")
}

func (s CSVFlag) Set(v string) error {
	*s.Values = strings.Split(v, ",")
	return nil
}

func (f *Flags) ParseFlags(args []string) {
	flags := flag.NewFlagSet("hc-bundler", flag.ExitOnError)
	flags.BoolVar(&f.Dryrun, "dryrun", false, "Performing a dry run will display all commands without executing them")
	flags.StringVar(&f.OS, "os", "auto", "Override operating system detection")
	flags.BoolVar(&f.Consul, "consul", false, "Run Consul diagnostics")
	flags.BoolVar(&f.Nomad, "nomad", false, "Run Nomad diagnostics")
	flags.BoolVar(&f.TFE, "tfe", false, "Run Terraform Enterprise diagnostics")
	flags.BoolVar(&f.Vault, "vault", false, "Run Vault diagnostics")
	flags.BoolVar(&f.AllProducts, "all", false, "Run all available product diagnostics")
	flags.Var(&CSVFlag{&f.Includes}, "includes", "files or directories to include (comma-separated, file-*-globbing available if 'wrapped-*-in-single-quotes')\ne.g. '/var/log/consul-*,/var/log/nomad-*'")
	flags.StringVar(&f.Outfile, "outfile", "support.tar.gz", "Output file name")
	flags.Parse(args)
}

// FIXME(mkcp): I'm not sure there's a lot of value for wrapping this assignment in a point receiver method. It's simpler
//  to set this value directly without mutating it when we create the agent.
func (a *Agent) start() {
	a.Start = time.Now()
}

func (a *Agent) end() {
	a.End = time.Now()
	a.Duration = fmt.Sprintf("%v seconds", a.End.Sub(a.Start).Seconds())
}

// CreateTemp Creates a temporary directory so that we may gather results and files before compressing the final
//   artifact.
func (a *Agent) CreateTemp() (err error) {
	if a.Dryrun {
		return nil
	}

	a.tmpDir, err = ioutil.TempDir("./", "temp")
	if err != nil {
		a.l.Error("Error creating temp directory", "name", hclog.Fmt("%s", a.tmpDir), "message", err)
		return err
	}
	a.l.Debug("Created temp directory", "name", hclog.Fmt("./%s", a.tmpDir))

	return nil
}

// Cleanup attempts to delete the contents of the tempdir when the diagnostics are done.
func (a Agent) Cleanup() (err error) {
	if a.Dryrun {
		return nil
	}

	a.l.Debug("Cleaning up temporary files")

	err = os.RemoveAll(a.tmpDir)
	if err != nil {
		a.l.Warn("Failed to clean up temp dir", "message", err)
	}
	return err
}

// CopyIncludes copies user-specified files over to our tempdir.
func (a Agent) CopyIncludes() (err error) {
	if len(a.Includes) == 0 {
		return nil
	}

	a.l.Info("Copying includes")

	dest := filepath.Join(a.tmpDir, "includes")
	if !a.Dryrun {
		err = os.MkdirAll(dest, 0755)
		if err != nil {
			return err
		}
	}

	for _, f := range a.Includes {
		if a.Dryrun {
			a.l.Info("Would include", "from", f)
			continue
		}
		a.l.Debug("getting Copier", "path", f)
		s := seeker.NewCopier(f, dest, false)
		if _, err = s.Run(); err != nil {
			return err
		}
	}

	return nil
}

// GetSeekers maps the products we'll inspect into the seekers that we'll execute.
func (a *Agent) GetSeekers() (err error) {
	a.l.Debug("Gathering Seekers")

	a.seekers, err = products.GetSeekers(a.Consul, a.Nomad, a.TFE, a.Vault, a.AllProducts, a.tmpDir)
	if err != nil {
		a.l.Error("products.GetSeekers", "error", err)
		return err
	}
	// TODO(kit): We need multiple independent seeker sets to execute these concurrently.
	a.seekers = append(a.seekers, hostdiag.NewHostSeeker(a.OS))
	a.NumSeekers = len(a.seekers)
	return nil
}

// RunSeekers executes all seekers for this run.
func (a *Agent) RunSeekers() (err error) {
	a.l.Info("Gathering diagnostics")

	err = a.GetSeekers()
	if err != nil {
		return err
	}

	// TODO(kit): Parallelize seeker set execution
	// TODO(kit): Extract the body of this loop out into a function?
	for _, s := range a.seekers {
		if a.Dryrun {
			a.l.Info("would run", "seeker", s.Identifier)
			continue
		}

		a.l.Info("running", "seeker", s.Identifier)
		a.results[s.Identifier] = s
		result, err := s.Run()
		if err != nil {
			a.NumErrors++
			a.l.Warn("result",
				"seeker", s.Identifier,
				"result", fmt.Sprintf("%s", result),
				"error", err,
			)
			if s.MustSucceed {
				a.l.Error("A critical Seeker failed", "message", err)
				return err
			}
		}
	}

	// TODO(kit): Users would benefit from us calculate the success rate here and always rendering it. Then we frame
	//  it as an error if we're over the 50% threshold. Maybe we only render it in the manifest or results?
	if a.NumErrors > a.NumSeekers/2 {
		return errors.New("more than 50% of Seekers failed")
	}

	return nil
}

// WriteOutput renders the manifest and results of the diagnostics run and writes the compressed archive.
func (a *Agent) WriteOutput() (err error) {
	a.end()

	if a.Dryrun {
		return nil
	}

	a.l.Debug("Writing results and manifest, and creating tar.gz archive")

	// Write out results
	rFile := filepath.Join(a.tmpDir, "Results.json")
	err = util.WriteJSON(a.results, rFile)
	if err != nil {
		a.l.Error("util.WriteJSON", "error", err)
		return err
	}
	a.l.Info("Created Results.json file", "dest", rFile)

	// Write out manifest
	mFile := filepath.Join(a.tmpDir, "Manifest.json")
	err = util.WriteJSON(a, mFile)
	if err != nil {
		a.l.Error("util.WriteJSON", "error", err)
		return err
	}
	a.l.Info("Created Manifest.json file", "dest", mFile)

	// Archive and compress outputs
	err = util.TarGz(a.tmpDir, a.Outfile)
	if err != nil {
		a.l.Error("util.TarGz", "error", err)
		return err
	}
	a.l.Info("Compressed and archived output file", "dest", a.Outfile)

	return nil
}
