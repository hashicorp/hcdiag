package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/host-diagnostics/hostdiag"
	"github.com/hashicorp/host-diagnostics/products"
	"github.com/hashicorp/host-diagnostics/seeker"
	"github.com/hashicorp/host-diagnostics/util"
)

var manifestOutputMap = map[string]interface{}{}

func main() {
	// TODO: standardize log and error handling
	// TODO: eval third party libs, gap and risk analysis
	// TODO: determine appropriate arguments, eval cli libs
	// TODO: hostdiag cmds and functions need some work, will be expanded based on initial feedback
	// TODO: allow multiple products - could use comma separate input and for each call to GetSeekers, or can handle in seekers func
	// TODO: validate temp dir cross platform
	// TODO: decide what exit codes we want with different error modes

	start := time.Now()
	dir := "."
	results := map[string]interface{}{}

	appLogger := configureLogging("host-diagnostics")

	// Parse arguments
	osPtr := flag.String("os", "auto", "(optional) Override operating system detection")
	productPtr := flag.String("product", "", "(optional) Run product diagnostic commands if specified")
	dryrunPtr := flag.Bool("dryrun", false, "(optional) Performing a dry run will display all commands without executing them")
	outfilePtr := flag.String("outfile", "support.tar.gz", "(optional) Output file name")
	flag.Parse()

	// dump flags to manifest (temporary)
	manifestOutputMap["OS"] = *osPtr
	manifestOutputMap["Product"] = *productPtr
	manifestOutputMap["Dryrun"] = *dryrunPtr
	manifestOutputMap["Outfile"] = *outfilePtr

	if !*dryrunPtr {
		// Create temporary directory for output files
		var err error
		dir, err = ioutil.TempDir("./", "temp")
		defer os.RemoveAll(dir)
		if err != nil {
			appLogger.Error("Error creating temp directory", "name", hclog.Fmt("%s", dir))
			os.Exit(1)
		}
		appLogger.Debug("Created temp directory", "name", hclog.Fmt("./%s", dir))

		defer writeOutput(manifestOutputMap, start, dir, &results, *outfilePtr)
	}

	appLogger.Info("Gathering diagnostics")
	// Set up Seekers
	seekers, err := products.GetSeekers(*productPtr, dir)
	if err != nil {
		appLogger.Error("products.GetSeekers", "error", err)
		os.Exit(1)
	}
	seekers = append(seekers, hostdiag.NewHostSeeker(*osPtr))
	manifestOutputMap["NumSeekers"] = len(seekers)

	// Run seekers
	results, err = RunSeekers(seekers, *dryrunPtr)
	if err != nil {
		appLogger.Error("a critical Seeker failed", "message", err)
		os.Exit(2)
	}
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
	errCount := 0

	for _, s := range seekers {
		if dry {
			l.Info("would run", "seeker", s.Identifier)
			continue
		}

		l.Info("running", "seeker", s.Identifier)
		results[s.Identifier] = s
		result, err := s.Run()
		if err != nil {
			errCount = errCount + 1
			l.Warn("result",
				"seeker", s.Identifier,
				"result", fmt.Sprintf("%s", result),
				"error", err,
			)
			if s.MustSucceed {
				manifestOutputMap["NumErrors"] = errCount
				return results, err
			}
		}
	}

	manifestOutputMap["NumErrors"] = errCount
	return results, nil
}

func writeOutput(manifestOutputMap map[string]interface{}, start time.Time, dir string, results *map[string]interface{}, outfile string) {
	l := hclog.Default()

	// Write out results
	err := util.WriteJSON(results, dir+"/Results.json")
	if err != nil {
		l.Error("util.WriteJSON", "error", err)
		os.Exit(1)
	}
	l.Info("created Results.json file", "dest", dir+"/Results.json")

	// Write manifest
	err = util.ManifestOutput(manifestOutputMap, start, dir)
	if err != nil {
		l.Error("util.ManifestOutput", "error", err)
		os.Exit(1)
	}
	l.Info("Created manifest output file", "dest", "./"+outfile)

	// Archive and compress outputs
	err = util.TarGz(dir, "./"+outfile)
	if err != nil {
		l.Error("util.TarGz", "error", err)
		os.Exit(1)
	}
	l.Info("Compressed and archived output file", "dest", "./"+outfile)
}
