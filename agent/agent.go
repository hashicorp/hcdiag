package agent

import (
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

	"github.com/hashicorp/hcdiag/seeker/host"

	"github.com/hashicorp/hcdiag/client"
	"github.com/hashicorp/hcdiag/version"
	"github.com/hashicorp/hcl/v2/hclsimple"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/product"
	"github.com/hashicorp/hcdiag/seeker"
	"github.com/hashicorp/hcdiag/util"
)

// Agent holds our set of seekers to be executed and their results.
type Agent struct {
	l           hclog.Logger
	products    map[string]*product.Product
	results     map[string]map[string]seeker.Seeker
	resultsLock sync.Mutex
	tmpDir      string
	Start       time.Time       `json:"started_at"`
	End         time.Time       `json:"ended_at"`
	Duration    string          `json:"duration"`
	NumSeekers  int             `json:"num_seekers"`
	Config      Config          `json:"configuration"`
	Version     version.Version `json:"version"`
	// ManifestSeekers holds a slice of seekers with a subset of normal seekers' fields so we can safely render them in
	// `manifest.json`
	ManifestSeekers map[string][]ManifestSeeker `json:"seekers"`
}

func NewAgent(config Config, logger hclog.Logger) *Agent {
	return &Agent{
		l:               logger,
		results:         make(map[string]map[string]seeker.Seeker),
		Config:          config,
		ManifestSeekers: make(map[string][]ManifestSeeker),
		Version:         version.GetVersion(),
	}
}

// Run manages the Agent's lifecycle. We create our temp directory, copy files, run their seekers, write the results,
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
	a.l.Debug("Gathering Products' Seekers")
	p, errProductSetup := a.Setup()
	if errProductSetup != nil {
		errs = append(errs, errProductSetup)
		a.l.Error("Failed running Products", "error", errProductSetup)
		return errs
	}

	// Filter the seekers on each product
	a.l.Debug("Applying Exclude and Select filters to products")
	for _, prod := range p {
		if errProductFilter := prod.Filter(); errProductFilter != nil {
			a.l.Error("Failed to filter Products", "error", errProductFilter)
			errs = append(errs, errProductFilter)
			return errs
		}
	}

	// Store products
	a.products = p

	// Sum up all seekers from products
	a.NumSeekers = product.CountSeekers(a.products)

	// Run products
	a.l.Info("Gathering diagnostics")
	if errProduct := a.RunProducts(); errProduct != nil {
		errs = append(errs, errProduct)
		a.l.Error("Failed running Products", "error", errProduct)
	}

	// Record metadata
	// Build seeker metadata
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

	a.l.Info("Writing summary of products and seekers to standard output")
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

	// glob "*" here is to support copy/paste of seeker identifiers
	// from -dryrun output into select/exclude filters
	a.tmpDir = "*"
	a.l.Info("Would copy included files", "includes", a.Config.Includes)

	// Running healthchecks for products. We don't want to stop if any fail though.
	a.l.Info("Checking product availability")
	if errProductChecks := a.CheckAvailable(); errProductChecks != nil {
		errs = append(errs, errProductChecks)
		a.l.Error("Product failed healthcheck. Ensure setup steps are complete (see https://github.com/hashicorp/hcdiag for prerequisites)", "error", errProductChecks)
	}

	// Create products and their seekers
	a.l.Info("Gathering seekers for each product")
	p, errProductSetup := a.Setup()
	if errProductSetup != nil {
		errs = append(errs, errProductSetup)
		a.l.Error("Failed gathering seekers for products", "error", errProductSetup)
		return errs
	}
	a.l.Info("Filtering seeker lists")
	for _, prod := range p {
		if errProductFilter := prod.Filter(); errProductFilter != nil {
			a.l.Error("Failed to filter Products", "error", errProductFilter)
			errs = append(errs, errProductFilter)
			return errs
		}
	}
	// TODO(mkcp): We should pass the products forward in run, not store them on the agent.
	a.products = p

	a.l.Info("Showing diagnostics that would be gathered")
	for _, p := range a.products {
		set := p.Seekers
		for _, s := range set {
			a.l.Info("would run", "product", p.Name, "seeker", s.Identifier)
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
		s := seeker.NewCopier(f, dest, a.Config.Since, a.Config.Until)
		if _, err = s.Run(); err != nil {
			return err
		}
	}

	return nil
}

// RunProducts executes all seekers for this run.
// TODO(mkcp): We can avoid locking and waiting on results if all results are generated async. Then they can get streamed
//  back to the dispatcher and merged into either a sync.Map or a purpose-built results map with insert(), read(), and merge().
func (a *Agent) RunProducts() error {
	// Set up our waitgroup to make sure we don't proceed until all products execute.
	wg := sync.WaitGroup{}
	wg.Add(len(a.products))

	// NOTE(mkcp): Wraps product.Run(), writes to the a.results map, then counts up the results
	run := func(wg *sync.WaitGroup, name string, product *product.Product) {
		result := product.Run()

		// Write results
		a.resultsLock.Lock()
		a.results[name] = result
		a.resultsLock.Unlock()

		statuses, err := seeker.StatusCounts(product.Seekers)
		if err != nil {
			a.l.Error("Error rendering seeker statuses", "product", product, "error", err)
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
	for name, p := range a.products {
		for _, s := range p.Seekers {
			m := ManifestSeeker{
				ID:     s.Identifier,
				Error:  s.ErrString,
				Status: s.Status,
			}
			a.ManifestSeekers[name] = append(a.ManifestSeekers[name], m)
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

func (a *Agent) Setup() (map[string]*product.Product, error) {
	// TODO(mkcp): Products.Config and agent ProductConfig is hella confusing and should be refactored
	// NOTE(mkcp): product.Config is a config struct with common params between product. while ProductConfig are the
	//  product-specific values we take in from HCL. Very confusing and needs work!
	var boundary, consul, nomad, tfe, vault, waypoint *ProductConfig

	for _, p := range a.Config.Products {
		switch p.Name {
		case product.Boundary:
			boundary = p
		case product.Consul:
			consul = p
		case product.Nomad:
			nomad = p
		case product.TFE:
			tfe = p
		case product.Vault:
			vault = p
		case product.Waypoint:
			waypoint = p
		}
	}

	cfg := product.Config{
		TmpDir:        a.tmpDir,
		Since:         a.Config.Since,
		Until:         a.Config.Until,
		OS:            a.Config.OS,
		DebugDuration: a.Config.DebugDuration,
		DebugInterval: a.Config.DebugInterval,
	}
	p := make(map[string]*product.Product)
	if a.Config.Boundary {
		newBoundary, err := product.NewBoundary(a.l, cfg)
		if err != nil {
			return nil, err
		}
		if boundary != nil {
			customSeekers, err := customSeekers(boundary, a.tmpDir)
			if err != nil {
				return nil, err
			}
			newBoundary.Seekers = append(newBoundary.Seekers, customSeekers...)
			newBoundary.Excludes = boundary.Excludes
			newBoundary.Selects = boundary.Selects
		}
		p[product.Boundary] = newBoundary

	}
	if a.Config.Consul {
		newConsul, err := product.NewConsul(a.l, cfg)
		if err != nil {
			return nil, err
		}
		if consul != nil {
			customSeekers, err := customSeekers(consul, a.tmpDir)
			if err != nil {
				return nil, err
			}
			newConsul.Seekers = append(newConsul.Seekers, customSeekers...)
			newConsul.Excludes = consul.Excludes
			newConsul.Selects = consul.Selects
		}
		p[product.Consul] = newConsul

	}
	if a.Config.Nomad {
		newNomad, err := product.NewNomad(a.l, cfg)
		if err != nil {
			return nil, err
		}
		if nomad != nil {
			customSeekers, err := customSeekers(nomad, a.tmpDir)
			if err != nil {
				return nil, err
			}
			newNomad.Seekers = append(newNomad.Seekers, customSeekers...)
			newNomad.Excludes = nomad.Excludes
			newNomad.Selects = nomad.Selects
		}
		p[product.Nomad] = newNomad
	}
	if a.Config.TFE {
		newTFE, err := product.NewTFE(a.l, cfg)
		if err != nil {
			return nil, err
		}
		if tfe != nil {
			customSeekers, err := customSeekers(tfe, a.tmpDir)
			if err != nil {
				return nil, err
			}
			newTFE.Seekers = append(newTFE.Seekers, customSeekers...)
			newTFE.Excludes = tfe.Excludes
			newTFE.Selects = tfe.Selects
		}
		p[product.TFE] = newTFE
	}
	if a.Config.Vault {
		newVault, err := product.NewVault(a.l, cfg)
		if err != nil {
			return nil, err
		}
		if vault != nil {
			customSeekers, err := customSeekers(vault, a.tmpDir)
			if err != nil {
				return nil, err
			}
			newVault.Seekers = append(newVault.Seekers, customSeekers...)
			newVault.Excludes = vault.Excludes
			newVault.Selects = vault.Selects
		}
		p[product.Vault] = newVault
	}
	if a.Config.Waypoint {
		newWaypoint, err := product.NewWaypoint(a.l, cfg)
		if err != nil {
			return nil, err
		}
		if vault != nil {
			customSeekers, err := customSeekers(waypoint, a.tmpDir)
			if err != nil {
				return nil, err
			}
			newWaypoint.Seekers = append(newWaypoint.Seekers, customSeekers...)
			newWaypoint.Excludes = vault.Excludes
			newWaypoint.Selects = vault.Selects
		}
		p[product.Vault] = newWaypoint
	}

	newHost := product.NewHost(a.l, cfg)
	if a.Config.Host != nil {
		customSeekers, err := customHostSeekers(a.Config.Host, a.tmpDir)
		if err != nil {
			return nil, err
		}
		newHost.Seekers = append(newHost.Seekers, customSeekers...)
		newHost.Excludes = a.Config.Host.Excludes
		newHost.Selects = a.Config.Host.Selects
	}
	p[product.Host] = newHost

	// product.Config is a config struct with common params between product
	return p, nil
}

func ParseHCL(path string) (Config, error) {
	// Parse our HCL
	var config Config
	err := hclsimple.DecodeFile(path, nil, &config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}

// WriteSummary writes a summary report that includes the products and seeker statuses present in the agent's
// ManifestSeekers. The intended use case is to write to output at the end of the Agent's Run.
func (a *Agent) WriteSummary(writer io.Writer) error {
	t := tabwriter.NewWriter(writer, 0, 0, 2, ' ', 0)
	headers := []string{
		"product",
		string(seeker.Success),
		string(seeker.Fail),
		string(seeker.Unknown),
		"total",
	}

	_, err := fmt.Fprint(t, formatReportLine(headers...))
	if err != nil {
		return err
	}

	// For deterministic output, we sort the products in alphabetical order. Otherwise, ranging over the map
	// a.ManifestSeekers directly, we wouldn't know for certain which order the keys - and therefore the rows - would be in.
	var products []string
	for k := range a.ManifestSeekers {
		products = append(products, k)
	}
	sort.Strings(products)

	for _, prod := range products {
		var success, fail, unknown int
		seekers := a.ManifestSeekers[prod]

		for _, s := range seekers {
			switch s.Status {
			case seeker.Success:
				success++
			case seeker.Fail:
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
			strconv.Itoa(len(seekers))))
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

// TODO(mkcp): This duplicates much of customSeekers and can certainly be improved.
func customHostSeekers(cfg *HostConfig, tmpDir string) ([]*seeker.Seeker, error) {
	seekers := make([]*seeker.Seeker, 0)
	// Build Commanders
	for _, c := range cfg.Commands {
		cmder := seeker.NewCommander(c.Run, c.Format)
		seekers = append(seekers, cmder)
	}
	// Build Shellers
	for _, c := range cfg.Shells {
		sheller := seeker.NewSheller(c.Run)
		seekers = append(seekers, sheller)
	}

	for _, g := range cfg.GETs {
		seekers = append(seekers, host.NewGetter(g.Path))
	}

	// Build copiers
	dest := tmpDir + "/host"
	for _, c := range cfg.Copies {
		var from time.Time

		// Set `from` with a timestamp
		if c.Since != "" {
			since, err := time.ParseDuration(c.Since)
			if err != nil {
				return nil, err
			}
			// Get the timestamp which marks the start of our duration
			from = time.Now().Add(-since)
		}

		copier := seeker.NewCopier(c.Path, dest, from, time.Time{})
		seekers = append(seekers, copier)
	}

	return seekers, nil
}

// TODO(mkcp): Products, not the agent, should handle their own custom seekers when they're created.
func customSeekers(cfg *ProductConfig, tmpDir string) ([]*seeker.Seeker, error) {
	seekers := make([]*seeker.Seeker, 0)
	// Build Commanders
	for _, c := range cfg.Commands {
		cmder := seeker.NewCommander(c.Run, c.Format)
		seekers = append(seekers, cmder)
	}
	// Build Shellers
	for _, c := range cfg.Shells {
		sheller := seeker.NewSheller(c.Run)
		seekers = append(seekers, sheller)
	}

	// Build HTTPers
	var c *client.APIClient
	var err error
	switch cfg.Name {
	case product.Boundary:
		c, err = client.NewBoundaryAPI()
	case product.Consul:
		c, err = client.NewConsulAPI()
	case product.Nomad:
		c, err = client.NewNomadAPI()
	case product.TFE:
		c, err = client.NewTFEAPI()
	case product.Vault:
		c, err = client.NewVaultAPI()
	case product.Waypoint:
		c, err = client.NewWaypointAPI()
	}
	if err != nil {
		return nil, err
	}
	for _, g := range cfg.GETs {
		httper := seeker.NewHTTPer(c, g.Path)
		seekers = append(seekers, httper)
	}

	// Build copiers
	dest := tmpDir + "/" + cfg.Name
	for _, c := range cfg.Copies {
		var from time.Time

		// Set `from` with a timestamp
		if c.Since != "" {
			since, err := time.ParseDuration(c.Since)
			if err != nil {
				return nil, err
			}
			// Get the timestamp which marks the start of our duration
			from = time.Now().Add(-since)
		}
		copier := seeker.NewCopier(c.Path, dest, from, time.Time{})
		seekers = append(seekers, copier)
	}

	return seekers, nil
}
