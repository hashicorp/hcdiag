package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/host-diagnostics/hostdiag"
	"github.com/hashicorp/host-diagnostics/products"
	"github.com/hashicorp/host-diagnostics/seeker"
	"github.com/hashicorp/host-diagnostics/util"
)

// TODO: NewDryAgent() to simplify all the 'if d.Dryrun's ??

func NewAgent(logger hclog.Logger) *Agent {
	a := Agent{
		l:       logger,
		results: make(map[string]map[string]interface{}),
	}
	return &a
}

// Agent holds our set of seekers to be executed and their results.
type Agent struct {
	l       hclog.Logger
	seekers map[string][]*seeker.Seeker
	results map[string]map[string]interface{}
	resultsLock sync.Mutex
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
	Serial      bool
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

func (f *Flags) ParseFlags(args []string) error {
	flags := flag.NewFlagSet("hc-bundler", flag.ExitOnError)
	flags.BoolVar(&f.Dryrun, "dryrun", false, "Performing a dry run will display all commands without executing them")
	flags.BoolVar(&f.Serial, "serial", false, "Run products in sequence rather than concurrently")
	flags.BoolVar(&f.Consul, "consul", false, "Run Consul diagnostics")
	flags.BoolVar(&f.Nomad, "nomad", false, "Run Nomad diagnostics")
	flags.BoolVar(&f.TFE, "tfe", false, "Run Terraform Enterprise diagnostics")
	flags.BoolVar(&f.Vault, "vault", false, "Run Vault diagnostics")
	flags.BoolVar(&f.AllProducts, "all", false, "Run all available product diagnostics")
	flags.StringVar(&f.OS, "os", "auto", "Override operating system detection")
	flags.StringVar(&f.Outfile, "outfile", "support.tar.gz", "Output file name")
	flags.Var(&CSVFlag{&f.Includes}, "includes", "files or directories to include (comma-separated, file-*-globbing available if 'wrapped-*-in-single-quotes')\ne.g. '/var/log/consul-*,/var/log/nomad-*'")

	return flags.Parse(args)
}


// Run manages the Agent's lifecycle. We create our temp directory, copy files, run their seekers, write the results,
// and finally cleanup after ourselves. Each step must run, so we collect any errors up and return them to the caller.
func (a *Agent) Run() []error {
	var errs []error

	// Begin execution, copy files and run seekers
	a.Start = time.Now()
	if errTemp := a.CreateTemp(); errTemp != nil {
		errs = append(errs, errTemp)
		a.l.Error("Failed to create temp directory", "error", errTemp)
	}
	if errCopy := a.CopyIncludes(); errCopy != nil {
		errs = append(errs, errCopy)
		a.l.Error("Failed copying includes", "error", errCopy)
	}
	if errSeeker := a.RunSeekers(); errSeeker != nil {
		errs = append(errs, errSeeker)
		a.l.Error("Failed running Seekers", "error", errSeeker)
	}

	// Execution finished, write our results and cleanup
	a.recordEnd()
	if errWrite := a.WriteOutput(); errWrite != nil {
		errs = append(errs, errWrite)
		a.l.Error("Failed running output", "error", errWrite)
	}
	if errCleanup := a.Cleanup(); errCleanup != nil {
		errs = append(errs, errCleanup)
		a.l.Error("Failed to cleanup after the run", "error", errCleanup)
	}
	return errs
}

func (a *Agent) recordEnd() {
	// Record the end timestamps so we can write it out.
	a.End = time.Now()
	a.Duration = fmt.Sprintf("%v seconds", a.End.Sub(a.Start).Seconds())
}

// CreateTemp Creates a temporary directory so that we may gather results and files before compressing the final
//   artifact.
func (a *Agent) CreateTemp() error {
	var err error
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
func (a *Agent) Cleanup() (err error) {
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
func (a *Agent) CopyIncludes() (err error) {
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
func (a *Agent) GetSeekers() error {
	var err error
	config := a.productConfig()

	a.l.Debug("Gathering Seekers")
	var seekers map[string][]*seeker.Seeker
	seekers, err = products.GetSeekers(config)
	if err != nil {
		a.l.Error("products.GetSeekers", "error", err)
		return err
	}
	seekers["host"] = hostdiag.GetSeekers(a.OS)
	// TODO(mkcp): We should probably write a merge function to union seeker sets together under their keys
	a.seekers = seekers
	a.NumSeekers = countSeekers(seekers)
	return nil
}

// RunSeekers executes all seekers for this run.
func (a *Agent) RunSeekers() error {
	var err error
	a.l.Info("Gathering diagnostics")

	err = a.GetSeekers()
	if err != nil {
		return err
	}

	// Set up our waitgroup to make sure we don't proceed until all products execute.
	wg := sync.WaitGroup{}
	wg.Add(len(a.seekers))

	// NOTE(mkcp): Create a closure around runSet and wg.Done(). This is a little complex, but saves us duplication
	//   in the product loop. Maybe we extract this to a private package function in the future?
	f := func(wg *sync.WaitGroup, product string, set []*seeker.Seeker) {
		result, err := a.runSet(product, set)
		a.resultsLock.Lock()
		a.results[product] = result
		a.resultsLock.Unlock()
		if err != nil {
			a.l.Error("Error running seekers", "product", product, "error", err)
		}
		wg.Done()
	}
	for product, set := range a.seekers {
		// Run synchronously if -serial is enabled
		if a.Serial {
			f(&wg, product, set)
			continue
		}
		// Run concurrently by default
		go func(product string, set []*seeker.Seeker) {
			f(&wg, product, set)
		}(product, set)
	}

	// Wait until every product is finished
	wg.Wait()

	// TODO(kit): Users would benefit from us calculate the success rate here and always rendering it. Then we frame
	//  it as an error if we're over the 50% threshold. Maybe we only render it in the manifest or results?
	if a.NumErrors > a.NumSeekers/2 {
		return errors.New("more than 50% of Seekers failed")
	}

	return nil
}

// WriteOutput renders the manifest and results of the diagnostics run and writes the compressed archive.
func (a *Agent) WriteOutput() (err error) {
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

// NOTE(mkcp): Not sure if this state -> config fn should be in the agent package. I don't love that it's behavior
//  on the agent struct
func (a *Agent) productConfig() products.Config {
	if a.AllProducts {
		config := products.NewConfigAllEnabled()
		config.TmpDir = a.tmpDir
		return config
	}
	return products.Config{
		Consul: a.Consul,
		Nomad:  a.Nomad,
		TFE:    a.TFE,
		Vault:  a.Vault,
		TmpDir: a.tmpDir,
	}
}

// runSeekers runs the seekers
// TODO(mkcp): Should we return a collection of errors from here?
func (a *Agent) runSet(product string, set []*seeker.Seeker) (map[string]interface{}, error) {
	a.l.Info("Running seekers for", "product", product)
	results := make(map[string]interface{})
	for _, s := range set  {
		if a.Dryrun {
			a.l.Info("would run", "seeker", s.Identifier)
			continue
		}

		a.l.Info("running", "seeker", s.Identifier)
		result, err := s.Run()
		results[s.Identifier] = s
		if err != nil {
			a.NumErrors++
			a.l.Warn("result",
				"seeker", s.Identifier,
				"result", fmt.Sprintf("%s", result),
				"error", err,
			)
			if s.MustSucceed {
				a.l.Error("A critical Seeker failed", "message", err)
				return results, err
			}
		}
	}
	return results, nil
}


// TODO(mkcp): probably put this in the seekers package
// TODO(mkcp): test this !
// countSeekers iterates over each set of seekers and sums their counts
func countSeekers(sets map[string][]*seeker.Seeker) int {
	var count int
	for _, s := range sets  {
		count = count + len(s)
	}
	return count
}
