package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/host-diagnostics/hostdiag"
	"github.com/hashicorp/host-diagnostics/products"
	"github.com/hashicorp/host-diagnostics/seeker"
	"github.com/hashicorp/host-diagnostics/util"
)

func main() {
	// TODO: standardize log and error handling
	// TODO: eval third party libs, gap and risk analysis
	// TODO: determine appropriate arguments, eval cli libs
	// TODO: update data model, lots of things generic currently
	// TODO: expand os and product cmds, os commands are really just placeholders atm
	// TODO: add support to targz for multiple files / dir, improve func; found several good examples but wanted to understand myself before using
	// TODO: expand hostdiag process, currently only returning all process names and not very useful
	// TODO: add outfile arg logic or similar, possibly options for output type
	// TODO: validate temp dir cross platform
	// TODO: decide what exit codes we want with different error modes

	appLogger := configureLogging("host-diagnostics")

	// Create temporary directory for output files
	dir, err := ioutil.TempDir("./", "temp")
	defer os.RemoveAll(dir)
	if err != nil {
		appLogger.Error("Error creating temp directory", "name", hclog.Fmt("%s", dir))
		os.Exit(1)
	}
	appLogger.Debug("Created temp directory", "name", hclog.Fmt("./%s", dir))

	// Parse arugments
	osPtr := flag.String("os", "auto", "(optional) Override operating system detection")
	productPtr := flag.String("product", "", "(optional) Run product diagnostic commands if specified")
	dryrunPtr := flag.Bool("dryrun", false, "(optional) Performing a dry run will display all commands without executing them")
	outfilePtr := flag.String("outfile", "support.tar.gz", "(optional) Output file name")
	flag.Parse()

	appLogger.Info("Gathering diagnostics")

	// set up Seekers
	seekers, err := products.GetSeekers(*productPtr, dir)
	if err != nil {
		appLogger.Error("products.GetSeekers", "error", err)
		os.Exit(1)
	}
	seekers = append(seekers, hostdiag.NewHostSeeker(*osPtr))

	// run em
	results, err := RunSeekers(seekers, *dryrunPtr)
	if err != nil {
		appLogger.Error("a critical Seeker failed", "message", err)
		os.Exit(2)
	}
	if *dryrunPtr {
		return
	}

	// write out results
	err = util.WriteJSON(results, dir+"/Results.json")
	if err != nil {
		os.Exit(1)
	}
	appLogger.Info("created Results.json file", "dest", dir+"/Results.json")

	// Create and compress archive of files in temporary folder
	appLogger.Info("Compressing and archiving output file", "name", *outfilePtr)
	util.TarGz(dir, "./"+*outfilePtr)
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
