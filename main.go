package main

import (
	"os"

	"github.com/hashicorp/go-hclog"
)

func main() {
	os.Exit(realMain())
}

// NOTE(mkcp): Most of the steps after parsing flags feel like they ought to be executed by an agent's run fn, not
//  orchestrated by main
func realMain() (returnCode int) {
	var err error
	l := configureLogging("host-diagnostics")
	a := NewAgent(l)
	if err := a.ParseFlags(os.Args[1:]); err != nil {
		// TODO(mkcp): Is there a specific return code for failing to parse input, or invalid input provided?
		return 1
	}
	if err := a.CreateTemp(); err != nil {
		// TODO(mkcp):  Is there a specific return code for failing to create a (temp) directory?
		return 1
	}

	// defer to ensure output-writing and cleanup even if there are seeker errors,
	// but update 'rc' so we can still exit non-0 on errors.
	// Cleanup being defer'd first makes it run last.
	defer func() {
		if err = a.Cleanup(); err != nil {
			returnCode = 1
		}
	}()
	defer func() {
		if err = a.WriteOutput(); err != nil {
			returnCode = 1
		}
	}()

	err = a.CopyIncludes()
	if err != nil {
		l.Error("Failed to copyIncludes", "message", err)
		return 1
	}

	err = a.RunSeekers()
	if err != nil {
		l.Error("Failed running Seekers", "message", err)
		return 1
	}

	return 0
}

// configureLogging takes a logger name, sets the default configuration, grabs the LOG_LEVEL from our ENV vars, and
//  returns a configured and usable logger.
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
