package main

import (
	"os"

	"github.com/hashicorp/go-hclog"
)

func main() {
	os.Exit(realMain())
}

func realMain() (rc int) {
	// TODO: standardize log and error handling
	// TODO: eval third party libs, gap and risk analysis
	// TODO: determine appropriate arguments, eval cli libs
	// TODO: hostdiag cmds and functions should be expanded
	// TODO: validate temp dir cross platform
	// TODO: decide what exit codes we want with different error modes

	var err error
	l := configureLogging("host-diagnostics")
	d := NewDiagnosticator(l)

	// defer to ensure output-writing and cleanup even if there are seeker errors,
	// but update 'rc' so we can still exit non-0 on errors.
	// Cleanup being defer'd first makes it run last.
	defer func() {
		if err = d.Cleanup(); err != nil {
			rc = 1
		}
	}()
	defer func() {
		if err = d.WriteOutput(); err != nil {
			rc = 1
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
