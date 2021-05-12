package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/host-diagnostics/hostdiag"
	"github.com/hashicorp/host-diagnostics/products"
	"github.com/hashicorp/host-diagnostics/seeker"
	"github.com/hashicorp/host-diagnostics/util"
)

// TODO: NewDryDiagnosticator() to simplify all the 'if d.Dryrun's ??
// NOTE(kit): ^ That division seems rly sensible to me. Going to do a ponder on how to divide these up.

// PLSFIX(kit): This can fail if the temp directory can't be built, we should return an error here
func NewDiagnosticator(logger hclog.Logger) (*Diagnosticator, error) {
	d := Diagnosticator{
		l:       logger,
		results: make(map[string]interface{}),
	}
	d.start()
	d.ParseFlags()
	// PLSFIX(kit): If we can't create a temp dir that should stop the world. Let's log it out and abort
	err := d.CreateTemp()
	if err != nil {
		logger.Error("Failed to create temp directory", "error", err)
		// NOTE(kit): Not sure if we ought to return
		return nil, err
	}
	return &d, nil
}

// PLSFIX(kit): add doccomment
// Diagnosticator holds our set of seekers to be executed and their results.
type Diagnosticator struct {
	l       hclog.Logger
	seekers []*seeker.Seeker
	results map[string]interface{}
	tmpDir  string

	Manifest
}

// PLSFIX(kit): add doccomment
// Manifest holds the metadata for a diagnostics run for rendering later.
type Manifest struct {
	Start      time.Time
	End        time.Time
	Duration   string
	NumErrors  int
	NumSeekers int

	Flags
}

// PLSFIX(kit): add doccomment
// Flags stores our CLI inputs.
type Flags struct {
	OS          string
	Dryrun      bool
	Consul      bool
	Nomad       bool
	TFE         bool
	Vault       bool
	AllProducts bool
	IncludeDir  string
	IncludeFile string
	Outfile     string
}

// PLSFIX(kit): add doccomment
// ParseFlags maps our CLI input flags to their storage location, name, defaults, and description.
func (f *Flags) ParseFlags() {
	flag.BoolVar(&f.Dryrun, "dryrun", false, "(optional) Performing a dry run will display all commands without executing them")
	flag.StringVar(&f.OS, "os", "auto", "(optional) Override operating system detection")
	flag.BoolVar(&f.Consul, "consul", false, "(optional) Run consul diagnostics")
	flag.BoolVar(&f.Nomad, "nomad", false, "(optional) Run nomad diagnostics")
	flag.BoolVar(&f.TFE, "tfe", false, "(optional) Run TFE/TFC diagnostics")
	flag.BoolVar(&f.Vault, "vault", false, "(optional) Run vault diagnostics")
	flag.BoolVar(&f.AllProducts, "all", false, "(optional) Run all available product diagnostics")
	flag.StringVar(&f.IncludeDir, "include-dir", "", "(optional) Include a directory in the bundle (e.g. logs)")
	flag.StringVar(&f.IncludeFile, "include-file", "", "(optional) Include a file in the bundle")
	flag.StringVar(&f.Outfile, "outfile", "support.tar.gz", "(optional) Output file name")
	flag.Parse()
}

func (d *Diagnosticator) start() {
	d.Start = time.Now()
}

func (d *Diagnosticator) end() {
	d.End = time.Now()
	d.Duration = fmt.Sprintf("%v seconds", d.End.Sub(d.Start).Seconds())
}

// PLSFIX(kit): add doccomment
// CreateTemp Creates a temporary directory so that we may gather results and files before compressing the final
//   artifact.
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

// PLSFIX(kit): add doccomment
// Cleanup ensures our temp directory contents are deleted when our diagnostics are done.
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

// PLSFIX(kit): add doccomment
// CopyIncludes copies the specified files over to our tempdir.
func (d *Diagnosticator) CopyIncludes() (err error) {
	// no sense trying anything else if no includes are.. included.
	if d.IncludeDir == "" && d.IncludeFile == "" {
		return nil
	}

	if d.Dryrun {
		if d.IncludeDir != "" {
			d.l.Info("Would include directory copy", "from", d.IncludeDir)
		}
		if d.IncludeFile != "" {
			d.l.Info("Would include file copy", "from", d.IncludeFile)
		}
		return nil
	}

	d.l.Info("Copying includes")

	dest := filepath.Join(d.tmpDir, "includes")
	err = os.MkdirAll(dest, 0755)
	if err != nil {
		return err
	}

	if d.IncludeDir != "" {
		if err = util.CopyDir(dest, d.IncludeDir); err != nil {
			return err
		}
	}
	if d.IncludeFile != "" {
		if err = util.CopyDir(dest, d.IncludeFile); err != nil {
			return err
		}
	}
	return nil
}

// PLSFIX(kit): add doccomment
// GetSeekers maps the products we're running into the seekers that we're executing.
func (d *Diagnosticator) GetSeekers() (err error) {
	d.l.Debug("Gathering Seekers")

	d.seekers, err = products.GetSeekers(d.Consul, d.Nomad, d.TFE, d.Vault, d.AllProducts, d.tmpDir)
	if err != nil {
		d.l.Error("products.GetSeekers", "error", err)
		return err
	}
	// TODO(kit): We need multiple independent seeker sets to execute these concurrently.
	d.seekers = append(d.seekers, hostdiag.NewHostSeeker(d.OS))
	d.NumSeekers = len(d.seekers)
	return nil
}

// PLSFIX(kit): add doccomment
// RunSeekers executes each set of seekers for the products in this run.
func (d *Diagnosticator) RunSeekers() (err error) {
	d.l.Info("Gathering diagnostics")

	err = d.GetSeekers()
	if err != nil {
		return err
	}

	// TODO(kit): Parallelize seeker set execution
	// TODO(kit): Extract the body of this loop out into a function?
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

	// TODO(kit): Users would benefit from us calculate the success rate here and always rendering it. Then we frame
	//  it as a warn if we're over the 50% threshold. Maybe we only render it in the manifest or results?
	if d.NumErrors > d.NumSeekers/2 {
		return errors.New("more than 50% of Seekers failed")
	}

	return nil
}

// PLSFIX(kit): doccomment
// WriteOutput renders the manifest and results of the diagnostics run.
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
