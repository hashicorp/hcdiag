package main

import (
	"os"

	"github.com/hashicorp/go-hclog"
)

func main() {
	os.Exit(realMain())
}

func realMain() (returnCode int) {
	// TODO(mkcp): rename to support-bundler
	l := configureLogging("host-diagnostics")
	a := NewAgent(l)

	// Parse inputs
	if err := a.ParseFlags(os.Args[1:]); err != nil {
		return 64
	}

	// Run the agent
	// NOTE(mkcp): Are there semantic returnCodes we can send based on the agent failure?
	errs := a.Run()
	if 0 < len(errs) {
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
