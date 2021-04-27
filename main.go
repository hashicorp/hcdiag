package main

import (
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

func main() {
	os.Exit(realMain())
}

func realMain() int {
	// TODO: standardize log and error handling
	// TODO: eval third party libs, gap and risk analysis
	// TODO: determine appropriate arguments, eval cli libs
	// TODO: hostdiag cmds and functions should be expanded
	// TODO: validate temp dir cross platform
	// TODO: decide what exit codes we want with different error modes

	var err error
	var manifest Manifest
	var seekers []*seeker.Seeker
	manifest.Start = time.Now()
	appLogger := configureLogging("host-diagnostics")
	results := map[string]interface{}{}
	dir := "."

	// Parse arguments
	osPtr := flag.String("os", "auto", "(optional) Override operating system detection")
	consulPtr := flag.Bool("consul", false, "(optional) Run consul diagnostics")
	nomadPtr := flag.Bool("nomad", false, "(optional) Run nomad diagnostics")
	vaultPtr := flag.Bool("vault", false, "(optional) Run vault diagnostics")
	allProductsPtr := flag.Bool("all", false, "(optional) Run all available product diagnostics")
	dryrunPtr := flag.Bool("dryrun", false, "(optional) Performing a dry run will display all commands without executing them")
	outfilePtr := flag.String("outfile", "support.tar.gz", "(optional) Output file name")
	// TODO: support more than one dir or file
	includeDir := flag.String("include-dir", "", "(optional) Include a directory in the bundle (e.g. logs)")
	includeFile := flag.String("include-file", "", "(optional) Include a file in the bundle")
	flag.Parse()

	manifest.OS = *osPtr
	manifest.Consul = *consulPtr
	manifest.Nomad = *nomadPtr
	manifest.Vault = *vaultPtr
	manifest.AllProducts = *allProductsPtr
	manifest.Dryrun = *dryrunPtr
	manifest.IncludeDir = *includeDir
	manifest.IncludeFile = *includeFile
	manifest.Outfile = *outfilePtr

	if !*dryrunPtr {
		// Create temporary directory for output files
		dir, err = ioutil.TempDir("./", "temp")
		defer os.RemoveAll(dir)
		if err != nil {
			appLogger.Error("Error creating temp directory", "name", hclog.Fmt("%s", dir))
			return 1
		}
		appLogger.Debug("Created temp directory", "name", hclog.Fmt("./%s", dir))

		defer writeOutput(&manifest, &seekers, &results, dir, *outfilePtr)
	}

	err = copyIncludes(filepath.Join(dir, "includes"), *includeDir, *includeFile)
	if err != nil {
		appLogger.Error("failed to copyIncludes", "message", err)
		return 1
	}

	appLogger.Info("Gathering diagnostics")

	// Set up Seekers
	seekers, err = products.GetSeekers(*consulPtr, *nomadPtr, *vaultPtr, *allProductsPtr, dir)
	if err != nil {
		appLogger.Error("products.GetSeekers", "error", err)
		os.Exit(1)
	}
	seekers = append(seekers, hostdiag.NewHostSeeker(*osPtr))
	manifest.NumSeekers = len(seekers)

	// Run seekers
	results, err = RunSeekers(seekers, *dryrunPtr)
	if err != nil {
		appLogger.Error("A critical Seeker failed", "message", err)
		return 2
	}

	return 0
}

// Manifest struct is used to retain high level runtime information.
type Manifest struct {
	Start       time.Time
	End         time.Time
	Duration    string
	NumErrors   int
	NumSeekers  int
	OS          string
	Dryrun      bool
	Consul      bool
	Nomad       bool
	Vault       bool
	AllProducts bool
	IncludeDir  string
	IncludeFile string
	Outfile     string
}

func configureLogging(loggerName string) hclog.Logger {
	// Create logger, set default and log level
	appLogger := hclog.New(&hclog.LoggerOptions{
		Name: loggerName,
	})
	hclog.SetDefault(appLogger)
	if logStr := os.Getenv("LOG_LEVEL"); logStr != "" {
		if level := hclog.LevelFromString(logStr); level != hclog.NoLevel {
			appLogger.SetLevel(level)
			appLogger.Debug("Logger configuration change", "LOG_LEVEL", hclog.Fmt("%s", logStr))
		}
	}
	return hclog.Default()
}

func RunSeekers(seekers []*seeker.Seeker, dry bool) (map[string]interface{}, error) {
	results := make(map[string]interface{})
	l := hclog.Default()

	for _, s := range seekers {
		if dry {
			l.Info("would run", "seeker", s.Identifier)
			continue
		}

		l.Info("running", "seeker", s.Identifier)
		results[s.Identifier] = s
		result, err := s.Run()
		if err != nil {
			l.Warn("result",
				"seeker", s.Identifier,
				"result", fmt.Sprintf("%s", result),
				"error", err,
			)
			if s.MustSucceed {
				return results, err
			}
		}
	}

	return results, nil
}

func writeOutput(manifest *Manifest, seekers *[]*seeker.Seeker, results *map[string]interface{}, dir string, outfile string) {
	l := hclog.Default()

	// Error summary
	for _, s := range *seekers {
		if s.Error != nil {
			manifest.NumErrors++
		}
	}

	// Manifest timing
	manifest.End = time.Now()
	manifest.Duration = fmt.Sprintf("%v seconds", manifest.End.Sub(manifest.Start).Seconds())

	// Write out results
	err := util.WriteJSON(results, dir+"/Results.json")
	if err != nil {
		l.Error("util.WriteJSON", "error", err)
		os.Exit(1)
	}
	l.Info("Created Results.json file", "dest", dir+"/Results.json")

	// Write out manifest
	err = util.WriteJSON(manifest, dir+"/Manifest.json")
	if err != nil {
		l.Error("util.WriteJSON", "error", err)
		os.Exit(1)
	}
	l.Info("Created Manifest.json file", "dest", dir+"/Manifest.json")

	// Archive and compress outputs
	err = util.TarGz(dir, "./"+outfile)
	if err != nil {
		l.Error("util.TarGz", "error", err)
		os.Exit(1)
	}
	l.Info("Compressed and archived output file", "dest", "./"+outfile)
}

func copyIncludes(to, dir, file string) (err error) {
	if dir == "" && file == "" {
		return nil
	}

	err = os.MkdirAll(to, 0755)
	if err != nil {
		return err
	}

	if dir != "" {
		if err = util.CopyDir(to, dir); err != nil {
			return err
		}
	}
	if file != "" {
		if err = util.CopyDir(to, file); err != nil {
			return err
		}
	}
	return nil
}
