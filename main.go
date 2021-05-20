package main

import (
	"os"

	"github.com/hashicorp/go-hclog"
)

func main() {
	os.Exit(realMain())
}

func realMain() (returnCode int) {
	var err error
	l := configureLogging("host-diagnostics")
	d := NewAgent(l)
	d.ParseFlags(os.Args[1:])
	if err := d.CreateTemp(); err != nil {
		return 1
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
