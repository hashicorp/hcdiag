package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/host-diagnostics/hostdiag"
	"github.com/hashicorp/host-diagnostics/products"
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

	configureLogging()
	appLogger := hclog.Default()

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

	// Get list of OS Commands to run along with their attribute identifier, arguments, and format
	appLogger.Info("Gathering host diagnostics")
	OSCommands := hostdiag.OSCommands(*osPtr)

	// Create map for host info and execute os commands
	HostInfo := make(map[string]interface{}, 0)
	HostInfo, _ = util.ExecuteCommands(OSCommands, *dryrunPtr)
	HostInfo["Host"], _ = hostdiag.GetHost()
	HostInfo["Disk"], _ = hostdiag.GetDisk()
	HostInfo["Memory"], _ = hostdiag.GetMemory()
	HostInfo["Processes"], _ = hostdiag.GetProcesses()
	HostInfo["Network"], _ = hostdiag.GetNetwork()

	// Host info map into JSON
	HostInfoJSON, _ := util.InterfaceToJSON(HostInfo)

	// Dump host info JSON into a results file
	util.JSONToFile(HostInfoJSON, dir+"/HostInfo.json")
	appLogger.Debug("Created output file", "name", hclog.Fmt("./%s", dir+"/HostInfo.json"))

	// Optional product commands
	if *productPtr != "" {
		// Get Product Commands to run along with their attribute identifier, arguments, and format
		appLogger.Info("Gathering product diagnostics")
		ProductCommands := products.ProductCommands(*productPtr, dir)

		// Execute product commands
		ProductInfo, err := util.ExecuteCommands(ProductCommands, *dryrunPtr)
		if err != nil {
			log.Fatalf("Error in ExecuteCommands: %v", err)
		}

		// Product info map into JSON
		ProductInfoJSON, _ := util.InterfaceToJSON(ProductInfo)

		// Dump product info JSON into a results_product file
		util.JSONToFile(ProductInfoJSON, dir+"/ProductInfo.json")
		appLogger.Debug("Created output file", "name", hclog.Fmt("./%s", dir+"/ProductInfo.json"))
	}

	// Create and compress archive of files in temporary folder
	appLogger.Info("Compressing and archiving output file", "name", *outfilePtr)
	util.TarGz(dir, "./"+*outfilePtr)
}

func configureLogging() {
	// Create logger, set default and log level
	appLogger := hclog.New(&hclog.LoggerOptions{
		Name: "host-diagnostics",
	})
	hclog.SetDefault(appLogger)
	if logStr := os.Getenv("LOG_LEVEL"); logStr != "" {
		if level := hclog.LevelFromString(logStr); level != hclog.NoLevel {
			appLogger.SetLevel(level)
			appLogger.Debug("Logger configuration change", "LOG_LEVEL", hclog.Fmt("%s", logStr))
		}
	}
}
