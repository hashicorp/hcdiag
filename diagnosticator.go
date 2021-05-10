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

// TODO: NewDryagnosticator() to simplify all the 'if d.Dryrun's ??

func NewDiagnosticator(logger hclog.Logger) *Diagnosticator {
	d := Diagnosticator{
		l:       logger,
		results: make(map[string]interface{}),
	}
	d.start()
	return &d
}

type Diagnosticator struct {
	l       hclog.Logger
	seekers []*seeker.Seeker
	results map[string]interface{}
	tmpDir  string

	Manifest
}

type Manifest struct {
	Start      time.Time
	End        time.Time
	Duration   string
	NumErrors  int
	NumSeekers int

	Flags
}

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
	if s.Values != nil {
		return strings.Join(*s.Values, ",")
	}
	return ""
}

func (s CSVFlag) Set(v string) error {
	*s.Values = strings.Split(v, ",")
	return nil
}

func (f *Flags) ParseFlags(args []string) {
	flags := flag.NewFlagSet("hc-diagnosticator", flag.ExitOnError)
	flags.BoolVar(&f.Dryrun, "dryrun", false, "Performing a dry run will display all commands without executing them")
	flags.StringVar(&f.OS, "os", "auto", "Override operating system detection")
	flags.BoolVar(&f.Consul, "consul", false, "Run Consul diagnostics")
	flags.BoolVar(&f.Nomad, "nomad", false, "Run Nomad diagnostics")
	flags.BoolVar(&f.TFE, "tfe", false, "Run Terraform Enterprise diagnostics")
	flags.BoolVar(&f.Vault, "vault", false, "Run Vault diagnostics")
	flags.BoolVar(&f.AllProducts, "all", false, "Run all available product diagnostics")
	flags.Var(&CSVFlag{&f.Includes}, "includes", "files or directories to include (comma-separated, file-*-globbing available if 'wrapped-*-in-single-quotes')\ne.g. '/var/log/consul-*,/var/log/nomad-*'")
	flags.StringVar(&f.Outfile, "outfile", "support.tar.gz", "Output file name (default: support.tar.gz)")
	flags.Parse(args)
}

func (d *Diagnosticator) start() {
	d.Start = time.Now()
}

func (d *Diagnosticator) end() {
	d.End = time.Now()
	d.Duration = fmt.Sprintf("%v seconds", d.End.Sub(d.Start).Seconds())
}

func (d *Diagnosticator) CreateTemp() (err error) {
	if d.Dryrun {
		return nil
	}

	d.tmpDir, err = ioutil.TempDir("./", "temp")
	if err != nil {
		d.l.Error("Error creating temp directory", "name", hclog.Fmt("%s", d.tmpDir), "message", err)
		return err
	}
	d.l.Debug("Created temp directory", "name", hclog.Fmt("./%s", d.tmpDir))

	return nil
}

func (d *Diagnosticator) Cleanup() (err error) {
	if d.Dryrun {
		return nil
	}

	d.l.Debug("Cleaning up temporary files")

	err = os.RemoveAll(d.tmpDir)
	if err != nil {
		d.l.Warn("Failed to clean up temp dir", "message", err)
	}
	return err
}

func (d *Diagnosticator) CopyIncludes() (err error) {
	if len(d.Includes) == 0 {
		return nil
	}

	d.l.Info("Copying includes")

	dest := filepath.Join(d.tmpDir, "includes")
	if !d.Dryrun {
		err = os.MkdirAll(dest, 0755)
		if err != nil {
			return err
		}
	}

	for _, f := range d.Includes {
		if d.Dryrun {
			d.l.Info("Would include", "from", f)
			continue
		}
		if err = util.CopyDir(dest, f); err != nil {
			return err
		}
	}

	return nil
}

func (d *Diagnosticator) GetSeekers() (err error) {
	d.l.Debug("Gathering Seekers")

	d.seekers, err = products.GetSeekers(d.Consul, d.Nomad, d.TFE, d.Vault, d.AllProducts, d.tmpDir)
	if err != nil {
		d.l.Error("products.GetSeekers", "error", err)
		return err
	}
	d.seekers = append(d.seekers, hostdiag.NewHostSeeker(d.OS))
	d.NumSeekers = len(d.seekers)
	return nil
}

func (d *Diagnosticator) RunSeekers() (err error) {
	d.l.Info("Gathering diagnostics")

	err = d.GetSeekers()
	if err != nil {
		return err
	}

	for _, s := range d.seekers {
		if d.Dryrun {
			d.l.Info("would run", "seeker", s.Identifier)
			continue
		}

		d.l.Info("running", "seeker", s.Identifier)
		d.results[s.Identifier] = s
		result, err := s.Run()
		if err != nil {
			d.NumErrors++
			d.l.Warn("result",
				"seeker", s.Identifier,
				"result", fmt.Sprintf("%s", result),
				"error", err,
			)
			if s.MustSucceed {
				d.l.Error("A critical Seeker failed", "message", err)
				return err
			}
		}
	}

	if d.NumErrors > d.NumSeekers/2 {
		return errors.New("more than 50% of Seekers failed")
	}

	return nil
}

func (d *Diagnosticator) WriteOutput() (err error) {
	d.end()

	if d.Dryrun {
		return nil
	}

	d.l.Debug("Writing results and manifest, and creating tar.gz archive")

	// Write out results
	rFile := filepath.Join(d.tmpDir, "Results.json")
	err = util.WriteJSON(d.results, rFile)
	if err != nil {
		d.l.Error("util.WriteJSON", "error", err)
		return err
	}
	d.l.Info("Created Results.json file", "dest", rFile)

	// Write out manifest
	mFile := filepath.Join(d.tmpDir, "Manifest.json")
	err = util.WriteJSON(d, mFile)
	if err != nil {
		d.l.Error("util.WriteJSON", "error", err)
		return err
	}
	d.l.Info("Created Manifest.json file", "dest", mFile)

	// Archive and compress outputs
	err = util.TarGz(d.tmpDir, d.Outfile)
	if err != nil {
		d.l.Error("util.TarGz", "error", err)
		return err
	}
	d.l.Info("Compressed and archived output file", "dest", d.Outfile)

	return nil
}
