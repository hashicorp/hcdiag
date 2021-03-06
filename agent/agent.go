package agent

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/hashicorp/hcdiag/hcl"

	"github.com/hashicorp/hcdiag/op"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/product"
	"github.com/hashicorp/hcdiag/runner"
	"github.com/hashicorp/hcdiag/util"
	"github.com/hashicorp/hcdiag/version"
)

// Config stores user input from the CLI and an HCL file.
type Config struct {
	// HCL stores configuration we receive from the custom config.
	HCL         hcl.HCL   `json:"hcl"`
	OS          string    `json:"operating_system"`
	Serial      bool      `json:"serial"`
	Dryrun      bool      `json:"dry_run"`
	Consul      bool      `json:"consul_enabled"`
	Nomad       bool      `json:"nomad_enabled"`
	TFE         bool      `json:"terraform_ent_enabled"`
	Vault       bool      `json:"vault_enabled"`
	Since       time.Time `json:"since"`
	Until       time.Time `json:"until"`
	Includes    []string  `json:"includes"`
	Destination string    `json:"destination"`

	// DebugDuration
	DebugDuration time.Duration `json:"debug_duration"`
	// DebugInterval
	DebugInterval time.Duration `json:"debug_interval"`
}

// Agent stores the runtime state that we use throughout the Agent's lifecycle.
type Agent struct {
	l           hclog.Logger
	products    map[product.Name]*product.Product
	results     map[product.Name]map[string]op.Op
	resultsLock sync.Mutex
	tmpDir      string
	Start       time.Time       `json:"started_at"`
	End         time.Time       `json:"ended_at"`
	Duration    string          `json:"duration"`
	NumOps      int             `json:"num_ops"`
	Config      Config          `json:"configuration"`
	Version     version.Version `json:"version"`
	// ManifestOps holds a slice of ops with a subset of fields so we can safely render them in `manifest.json`
	ManifestOps map[string][]ManifestOp `json:"ops"`
}

func NewAgent(config Config, logger hclog.Logger) *Agent {
	return &Agent{
		l:           logger,
		Config:      config,
		results:     make(map[product.Name]map[string]op.Op),
		products:    make(map[product.Name]*product.Product),
		ManifestOps: make(map[string][]ManifestOp),
		Version:     version.GetVersion(),
	}
}

// Run manages the Agent's lifecycle. We create our temp directory, copy files, run their ops, write the results,
// and finally cleanup after ourselves. We collect any errors up and return them to the caller, only returning when done
// or if the error warrants ending the run early.
func (a *Agent) Run() []error {
	var errs []error

	a.Start = time.Now()

	// If dryrun is enabled we short circuit the main lifecycle and run the dryrun mode instead.
	if a.Config.Dryrun {
		return a.DryRun()
	}

	a.l.Info("Ensuring destination directory exists", "directory", a.Config.Destination)
	errDest := util.EnsureDirectory(a.Config.Destination)
	if errDest != nil {
		errs = append(errs, errDest)
		a.l.Error("Failed to ensure destination directory exists", "dir", a.Config.Destination, "error", errDest)
	}

	// File processing
	if errTemp := a.CreateTemp(); errTemp != nil {
		errs = append(errs, errTemp)
		a.l.Error("Failed to create temp directory", "error", errTemp)
	}
	if errCopy := a.CopyIncludes(); errCopy != nil {
		errs = append(errs, errCopy)
		a.l.Error("Failed copying includes", "error", errCopy)
	}

	// Product processing
	// If any of the products' healthchecks fail, we abort the run. We want to abort the run here so we don't encourage
	// users to send us incomplete diagnostics.
	a.l.Info("Checking product availability")
	if errProductChecks := a.CheckAvailable(); errProductChecks != nil {
		errs = append(errs, errProductChecks)
		a.l.Error("Failed Product Checks", "error", errProductChecks)
		// End the run if any product fails its checks.
		return errs
	}

	// Create products
	a.l.Debug("Gathering Products' Runners")
	errProductSetup := a.Setup()
	if errProductSetup != nil {
		errs = append(errs, errProductSetup)
		a.l.Error("Failed running Products", "error", errProductSetup)
		return errs
	}

	// Filter the ops on each product
	a.l.Debug("Applying Exclude and Select filters to products")
	for _, prod := range a.products {
		if errProductFilter := prod.Filter(); errProductFilter != nil {
			a.l.Error("Failed to filter Products", "error", errProductFilter)
			errs = append(errs, errProductFilter)
			return errs
		}
	}

	// Sum up all runners from products
	a.NumOps = product.CountRunners(a.products)

	// Run products
	a.l.Info("Gathering diagnostics")
	if errProduct := a.RunProducts(); errProduct != nil {
		errs = append(errs, errProduct)
		a.l.Error("Failed running Products", "error", errProduct)
	}

	// Record metadata
	// Build op metadata
	a.l.Info("Recording manifest")
	a.RecordManifest()

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

	a.l.Info("Writing summary of products and ops to standard output")
	if errSummary := a.WriteSummary(os.Stdout); errSummary != nil {
		errs = append(errs, errSummary)
		a.l.Error("Failed to write summary report following run", "error", errSummary)
	}
	return errs
}

// DryRun runs the agent to log what would occur during a run, without issuing any commands or writing to disk.
func (a *Agent) DryRun() []error {
	var errs []error

	a.l.Info("Starting dry run")

	// glob "*" here is to support copy/paste of runner identifiers
	// from -dryrun output into select/exclude filters
	a.tmpDir = "*"
	a.l.Info("Would copy included files", "includes", a.Config.Includes)

	// Running healthchecks for products. We don't want to stop if any fail though.
	a.l.Info("Checking product availability")
	if errProductChecks := a.CheckAvailable(); errProductChecks != nil {
		errs = append(errs, errProductChecks)
		a.l.Error("Product failed healthcheck. Ensure setup steps are complete (see https://github.com/hashicorp/hcdiag for prerequisites)", "error", errProductChecks)
	}

	// Create products and their ops
	a.l.Info("Gathering operations for each product")
	errProductSetup := a.Setup()
	if errProductSetup != nil {
		errs = append(errs, errProductSetup)
		a.l.Error("Failed gathering ops for products", "error", errProductSetup)
		return errs
	}
	a.l.Info("Filtering runner lists")
	for _, prod := range a.products {
		if errProductFilter := prod.Filter(); errProductFilter != nil {
			a.l.Error("Failed to filter Products", "error", errProductFilter)
			errs = append(errs, errProductFilter)
			return errs
		}
	}

	a.l.Info("Showing diagnostics that would be gathered")
	for _, p := range a.products {
		set := p.Runners
		for _, r := range set {
			a.l.Info("would run", "product", p.Name, "runner", r.ID())
		}
	}
	a.l.Info("Would write output", "dest", a.Config.Destination)
	a.l.Info("Dry run complete", "duration", time.Since(a.Start))
	return errs
}

func (a *Agent) recordEnd() {
	// Record the end timestamps so we can write it out.
	a.End = time.Now()
	a.Duration = fmt.Sprintf("%v seconds", a.End.Sub(a.Start).Seconds())
}

// TempDir returns "hcdiag-" and an ISO 8601-formatted timestamp for temporary directory and tar file names.
// e.g. "hcdiag-2021-11-22T223938Z"
func (a *Agent) TempDir() string {
	// specifically excluding colons ":" since they are anathema to some filesystems and programs.
	ts := a.Start.UTC().Format("2006-01-02T150405Z")
	return "hcdiag-" + ts
}

// CreateTemp Creates a temporary directory so that we may gather results and files before compressing the final
//  artifact.
func (a *Agent) CreateTemp() error {
	tmp, err := os.MkdirTemp(a.Config.Destination, a.TempDir())
	if err != nil {
		a.l.Error("Error creating temp directory", "message", err)
		return err
	}
	tmp, err = filepath.Abs(tmp)
	if err != nil {
		a.l.Error("Error identifying absolute path for temp directory", "message", err)
		return err
	}
	a.tmpDir = tmp
	a.l.Debug("Created temp directory", "name", a.tmpDir)

	return nil
}

// Cleanup attempts to delete the contents of the tempdir when the diagnostics are done.
func (a *Agent) Cleanup() (err error) {
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
	err = os.MkdirAll(dest, 0755)
	if err != nil {
		return err
	}

	for _, f := range a.Config.Includes {
		a.l.Debug("validating include exists", "path", f)
		// Wildcards don't represent a single file or dir, so we can't validate that they exist.
		if !strings.Contains(f, "*") {
			_, err := os.Stat(f)
			if err != nil {
				return err
			}
		}

		a.l.Debug("getting Copier", "path", f)
		o := runner.NewCopier(f, dest, a.Config.Since, a.Config.Until).Run()
		if o.Error != nil {
			return o.Error
		}
	}

	return nil
}

// RunProducts executes all ops for this run.
// TODO(mkcp): We can avoid locking and waiting on results if all results are generated async. Then they can get streamed
//  back to the dispatcher and merged into either a sync.Map or a purpose-built results map with insert(), read(), and merge().
func (a *Agent) RunProducts() error {
	// Set up our waitgroup to make sure we don't proceed until all products execute.
	wg := sync.WaitGroup{}
	wg.Add(len(a.products))

	// NOTE(mkcp): Wraps product.Run(), writes to the a.results map, then counts up the results
	run := func(wg *sync.WaitGroup, name product.Name, product *product.Product) {
		result := product.Run()

		// Write results
		a.resultsLock.Lock()
		a.results[name] = result
		a.resultsLock.Unlock()

		statuses, err := op.StatusCounts(result)
		if err != nil {
			a.l.Error("Error rendering op statuses", "product", product, "error", err)
		}

		a.l.Info("Product done", "product", name, "statuses", statuses)
		wg.Done()
	}

	// Run each product
	for name, p := range a.products {
		// Run synchronously if -serial is enabled
		if a.Config.Serial {
			run(&wg, name, p)
			continue
		}
		// Run concurrently by default
		go run(&wg, name, p)
	}

	// Wait until every product is finished
	wg.Wait()

	return nil
}

// RecordManifest writes additional data to the agent to serialize into manifest.json
func (a *Agent) RecordManifest() {
	for name, ops := range a.results {
		for _, o := range ops {
			m := ManifestOp{
				ID:     o.Identifier,
				Error:  o.ErrString,
				Status: o.Status,
			}
			a.ManifestOps[string(name)] = append(a.ManifestOps[string(name)], m)
		}
	}
}

// WriteOutput renders the manifest and results of the diagnostics run and writes the compressed archive.
func (a *Agent) WriteOutput() (err error) {
	err = util.EnsureDirectory(a.Config.Destination)
	if err != nil {
		a.l.Error("Failed to ensure destination directory exists", "dir", a.Config.Destination, "error", err)
		return err
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

	// Build bundle destination path from config
	resultsFile := fmt.Sprintf("%s.tar.gz", a.TempDir())
	resultsDest := filepath.Join(a.Config.Destination, resultsFile)

	// Archive and compress outputs
	err = util.TarGz(a.tmpDir, resultsDest, a.TempDir())
	if err != nil {
		a.l.Error("util.TarGz", "error", err)
		return err
	}
	a.l.Info("Compressed and archived output file", "dest", resultsDest)

	return nil
}

// CheckAvailable runs healthchecks for each enabled product
func (a *Agent) CheckAvailable() error {
	if a.Config.Consul {
		err := product.CommanderHealthCheck(product.ConsulClientCheck, product.ConsulAgentCheck)
		if err != nil {
			return err
		}
	}
	if a.Config.Nomad {
		err := product.CommanderHealthCheck(product.NomadClientCheck, product.NomadAgentCheck)
		if err != nil {
			return err
		}
	}
	// NOTE(mkcp): We don't have a TFE healthcheck because we don't support API checks yet.
	// if cfg.TFE {
	// }
	if a.Config.Vault {
		err := product.CommanderHealthCheck(product.VaultClientCheck, product.VaultAgentCheck)
		if err != nil {
			return err
		}
	}
	return nil
}

// Setup instantiates each enabled product for the run. We compose product-specific configuration with the shared config
// that each product needs to build its runners.
func (a *Agent) Setup() error {
	if a.products == nil {
		return errors.New("agent.products is nil")
	}

	// Convert the slice of HCL products into a map we can read entries from directly
	hclProducts := hcl.ProductsMap(a.Config.HCL.Products)

	// Create the base config that we copy into each product
	baseCfg := product.Config{
		TmpDir:        a.tmpDir,
		Since:         a.Config.Since,
		Until:         a.Config.Until,
		OS:            a.Config.OS,
		DebugDuration: a.Config.DebugDuration,
		DebugInterval: a.Config.DebugInterval,
	}

	// Build Consul and assign it to the product map.
	if a.Config.Consul {
		cfg := baseCfg
		cfg.HCL = hclProducts["consul"]
		newConsul, err := product.NewConsul(a.l, cfg)
		if err != nil {
			return err
		}
		a.products[product.Consul] = newConsul
	}
	// Build Nomad and assign it to the product map.
	if a.Config.Nomad {
		cfg := baseCfg
		cfg.HCL = hclProducts["nomad"]
		newNomad, err := product.NewNomad(a.l, cfg)
		if err != nil {
			return err
		}
		a.products[product.Nomad] = newNomad
	}
	// Build TFE and assign it to the product map.
	if a.Config.TFE {
		cfg := baseCfg
		cfg.HCL = hclProducts["terraform-ent"]
		newTFE, err := product.NewTFE(a.l, cfg)
		if err != nil {
			return err
		}
		a.products[product.TFE] = newTFE
	}
	// Build Vault and assign it to the product map.
	if a.Config.Vault {
		cfg := baseCfg
		cfg.HCL = hclProducts["vault"]
		newVault, err := product.NewVault(a.l, cfg)
		if err != nil {
			return err
		}
		a.products[product.Vault] = newVault
	}

	// Build host and assign it to the product map.
	newHost, err := product.NewHost(a.l, baseCfg, a.Config.HCL.Host)
	if err != nil {
		return err
	}
	a.products[product.Host] = newHost

	return nil
}

// WriteSummary writes a summary report that includes the products and op statuses present in the agent's
// ManifestOps. The intended use case is to write to output at the end of the Agent's Run.
func (a *Agent) WriteSummary(writer io.Writer) error {
	t := tabwriter.NewWriter(writer, 0, 0, 2, ' ', 0)
	headers := []string{
		"product",
		string(op.Success),
		string(op.Fail),
		string(op.Unknown),
		"total",
	}

	_, err := fmt.Fprint(t, formatReportLine(headers...))
	if err != nil {
		return err
	}

	// For deterministic output, we sort the products in alphabetical order. Otherwise, ranging over the map
	// a.ManifestOps directly, we wouldn't know for certain which order the keys - and therefore the rows - would be in.
	var products []string
	for k := range a.ManifestOps {
		products = append(products, k)
	}
	sort.Strings(products)

	for _, prod := range products {
		var success, fail, unknown int
		ops := a.ManifestOps[prod]

		for _, o := range ops {
			switch o.Status {
			case op.Success:
				success++
			case op.Fail:
				fail++
			default:
				unknown++
			}
		}

		_, err := fmt.Fprint(t, formatReportLine(
			prod,
			strconv.Itoa(success),
			strconv.Itoa(fail),
			strconv.Itoa(unknown),
			strconv.Itoa(len(ops))))
		if err != nil {
			return err
		}
	}

	err = t.Flush()
	if err != nil {
		return err
	}

	return err
}

func formatReportLine(cells ...string) string {
	format := ""

	// The coercion from the argument of type []string to type []interface is required for the later
	// call to fmt.Sprintf, in which variadic arguments must be of type any/interface{}.
	strValues := make([]interface{}, len(cells))
	for i, cell := range cells {
		format += "%s\t"
		strValues[i] = cell
	}

	format += "\n"

	return fmt.Sprintf(format, strValues...)
}
