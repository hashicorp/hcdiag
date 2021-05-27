package agent

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/host-diagnostics/hostdiag"
	"github.com/hashicorp/host-diagnostics/products"
	"github.com/hashicorp/host-diagnostics/seeker"
	"github.com/hashicorp/host-diagnostics/util"
)

// TODO: NewDryAgent() to simplify all the 'if d.Dryrun's ??

func NewAgent(config Config, logger hclog.Logger) *Agent {
	a := Agent{
		l:       logger,
		results: make(map[string]map[string]interface{}),
		Config:  config,
	}
	return &a
}

// Agent holds our set of seekers to be executed and their results.
type Agent struct {
	l           hclog.Logger
	seekers     map[string][]*seeker.Seeker
	results     map[string]map[string]interface{}
	resultsLock sync.Mutex
	tmpDir      string
	Start       time.Time `json:"started_at"`
	End         time.Time `json:"ended_at"`
	Duration    string    `json:"duration"`
	NumErrors   int       `json:"num_errors"`
	NumSeekers  int       `json:"num_seekers"`
	Config      Config    `json:"configuration"`
}

type Config struct {
	OS          string   `json:"operating_system"`
	Serial      bool     `json:"serial"`
	Dryrun      bool     `json:"dry_run"`
	Consul      bool     `json:"consul_enabled"`
	Nomad       bool     `json:"nomad_enabled"`
	TFE         bool     `json:"terraform_ent_enabled"`
	Vault       bool     `json:"vault_enabled"`
	AllProducts bool     `json:"all_products_enabled"`
	Includes    []string `json:"includes"`
	Outfile     string   `json:"out_file"`
}

// Run manages the Agent's lifecycle. We create our temp directory, copy files, run their seekers, write the results,
// and finally cleanup after ourselves. We collect any errors up and return them to the caller, only returning when done
// or if the error warrants ending the run early.
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
	pConfig := a.productConfig()
	if errProductChecks := a.CheckProducts(pConfig); errProductChecks != nil {
		errs = append(errs, errProductChecks)
		a.l.Error("Failed Product Checks", "error", errProductChecks)
		// End the run if any product fails its checks.
		return errs
	}
	if errProduct := a.RunProducts(pConfig); errProduct != nil {
		errs = append(errs, errProduct)
		a.l.Error("Failed running Products", "error", errProduct)
	}

	// Execution finished, write our results and cleanup
	a.recordEnd()
	if errWrite := a.WriteOutput(a.DestinationFileName()); errWrite != nil {
		errs = append(errs, errWrite)
		a.l.Error("Failed running output", "error", errWrite)
	}
	if errCleanup := a.Cleanup(); errCleanup != nil {
		errs = append(errs, errCleanup)
		a.l.Error("Failed to cleanup after the run", "error", errCleanup)
	}
	return errs
}

func (a *Agent) CheckProducts(config products.Config) error {
	// If any of the products' healthchecks fail, we abort the run. We want to abort the run here so we don't encourage
	// users to send us incomplete diagnostics.
	a.l.Info("Checking product availability")
	return products.CheckAvailable(config)
}

func (a *Agent) recordEnd() {
	// Record the end timestamps so we can write it out.
	a.End = time.Now()
	a.Duration = fmt.Sprintf("%v seconds", a.End.Sub(a.Start).Seconds())
}

// CreateTemp Creates a temporary directory so that we may gather results and files before compressing the final
//  artifact.
func (a *Agent) CreateTemp() error {
	var err error
	if a.Config.Dryrun {
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
	if a.Config.Dryrun {
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
	if len(a.Config.Includes) == 0 {
		return nil
	}

	a.l.Info("Copying includes")

	dest := filepath.Join(a.tmpDir, "includes")
	if !a.Config.Dryrun {
		err = os.MkdirAll(dest, 0755)
		if err != nil {
			return err
		}
	}

	for _, f := range a.Config.Includes {
		if a.Config.Dryrun {
			a.l.Info("Would include", "from", f)
			continue
		}
		a.l.Debug("getting Copier", "path", f)
		s := seeker.NewCopier(f, dest)
		if _, err = s.Run(); err != nil {
			return err
		}
	}

	return nil
}

// GetSeekers maps the products we'll inspect into the seekers that we'll execute.
func (a *Agent) GetProductSeekers(config products.Config) error {
	a.l.Debug("Gathering Products' Seekers")
	var err error
	var seekers map[string][]*seeker.Seeker
	seekers, err = products.GetSeekers(config)
	if err != nil {
		a.l.Error("agent.GetProductSeekers", "error", err)
		return err
	}
	seekers["host"] = hostdiag.GetSeekers(a.Config.OS)
	a.seekers = seekers
	a.NumSeekers = seeker.CountInSets(seekers)
	return nil
}

// RunSeekers executes all seekers for this run.
func (a *Agent) RunProducts(config products.Config) error {
	var err error
	a.l.Info("Gathering diagnostics")

	err = a.GetProductSeekers(config)
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
		if a.Config.Serial {
			f(&wg, product, set)
			continue
		}
		// Run concurrently by default
		go f(&wg, product, set)
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
func (a *Agent) WriteOutput(resultsDest string) (err error) {
	if a.Config.Dryrun {
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
	err = util.TarGz(a.tmpDir, resultsDest)
	if err != nil {
		a.l.Error("util.TarGz", "error", err)
		return err
	}
	a.l.Info("Compressed and archived output file", "dest", a.Config.Outfile)

	return nil
}

// runSeekers runs the seekers
// TODO(mkcp): Should we return a collection of errors from here?
func (a *Agent) runSet(product string, set []*seeker.Seeker) (map[string]interface{}, error) {
	a.l.Info("Running seekers for", "product", product)
	results := make(map[string]interface{})
	for _, s := range set {
		if a.Config.Dryrun {
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
		}
	}
	return results, nil
}

func (a *Agent) productConfig() products.Config {
	if a.Config.AllProducts {
		return products.NewConfigAllEnabled()
	}
	return products.Config{
		Consul: a.Config.Consul,
		Nomad:  a.Config.Nomad,
		TFE:    a.Config.TFE,
		Vault:  a.Config.Vault,
		TmpDir: a.tmpDir,
	}
}

// DestinationFileName appends an ISO 8601-formatted timestamp to the outfile name.
func (a *Agent) DestinationFileName() string {
	timestamp := time.Now().Format(time.RFC3339)
	return fmt.Sprintf("%s-%s.tar.gz", a.Config.Outfile, timestamp)
}
