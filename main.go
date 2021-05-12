package main

import (
	"os"

	"github.com/hashicorp/go-hclog"
)

func main() {
	os.Exit(realMain())
}

// PLSFIX(kit): Small nit, and I hate to be anti-fun again but let's call this _main() or run().
// PLSFIX(kit): expand `rc` to its full `returnCode`
func realMain() (returnCode int) {
	// PLXFIX(kit): Export this to gh issues
	// TODO: standardize log and error handling

	// TODO: eval third party libs, gap and risk analysis

	// PLXFIX(kit): This one is mostly done I think? We should create an issue
	//              that outlines what CLI args we can continue to add.
	// TODO: determine appropriate arguments, eval cli libs

	// PLXFIX(kit): Export this to gh issues
	//              Let's take a look at the initial scope in the tickets and
	//              compare it to where we're at now.
	// TODO: hostdiag cmds and functions should be expanded

	// PLXFIX(kit): Export this to gh issues
	//              and write a few test cases for it?
	// TODO: validate temp dir cross platform

	// PLXFIX(kit): Export this to gh issues
	// TODO: decide what exit codes we want with different error modes

	var err error
	l := configureLogging("host-diagnostics")
	d, err := NewDiagnosticator(l)
	// PLSFIX(kit): Stop the world if we get an error while creating our diag agent
	if err != nil {
		l.Error("Failed to create Diagnostics agent", "error", err)
		returnCode = 1
		return returnCode
	}

	// defer to ensure output-writing and cleanup even if there are seeker errors,
	// but update 'rc' so we can still exit non-0 on errors.
	// Cleanup being defer'd first makes it run last.
	defer func() {
		if err = d.Cleanup(); err != nil {
			returnCode = 1
		}
	}()
	defer func() {
		if err = d.WriteOutput(); err != nil {
			returnCode = 1
		}
	}()

	err = d.CopyIncludes()
	if err != nil {
		l.Error("Failed to copyIncludes", "message", err)
		return 1
	}

	err = d.RunSeekers()
	if err != nil {
		l.Error("Failed running Seekers", "message", err)
		return 1
	}

	return 0
}

// PLSFIX(kit): when all the config is tied together it turns into some non-trivial behavior here. let's add a doccomment
// configureLogging takes a logger name, sets the default configuration, grabs
//  the LOG_LEVEL from our ENV vars, and returns a configured and usable logger.
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
