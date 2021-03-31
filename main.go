package main

import (
	"flag"
	"io/ioutil"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/host-diagnostics/hostdiag"
	"github.com/hashicorp/host-diagnostics/products"
	"github.com/hashicorp/host-diagnostics/util"
)

func main() {
	// TODO: standardize error handling
	// TODO: eval third party libs, gap and risk analysis
	// TODO: determine appropriate arguments, eval cli libs
	// TODO: update data model, lots of things generic currently
	// TODO: expand os and product cmds, os commands are really just placeholders atm
	// TODO: add support to targz for multiple files / dir, improve func; found several good examples but wanted to understand myself before using
	// TODO: expand hostdiag process, currently only returning all process names and not very useful
	// TODO: add outfile arg logic or similar, possibly options for output type
	// TODO: consolidate output into single targz
	// TODO: validate temp dir cross platform

	// Create logger and set as default
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

	// Create temporary directory for output files
	dir, err := ioutil.TempDir("./", "temp")
	defer os.RemoveAll(dir)
	if err != nil {
		appLogger.Error("Error creating temp directory", "name", hclog.Fmt("%s", dir))
		os.Exit(1)
	}
	appLogger.Debug("Created temp directory", "name", hclog.Fmt("./%s", dir))

	// args
	osPtr := flag.String("os", "auto", "(optional) Override operating system detection")
	productPtr := flag.String("product", "", "(optional) Run product diagnostic commands if specified")
	dryrunPtr := flag.Bool("dryrun", false, "(optional) Performing a dry run will display all commands without executing them")
	outfilePtr := flag.String("outfile", "support.tar.gz", "(optional) Output file name")
	// levelPtr := flag.String("level", "full", "(optional) info gathering level") // basic, enhanced, full ??
	flag.Parse()

	appLogger.Info("Gathering host diagnostics")
	// Get list of OS Commands to run along with their attribute identifier, arguments, and format
	OSCommands := make([]util.CommandStruct, 0)
	OSCommands = hostdiag.OSCommands(*osPtr)

	// Create map for host info and execute os commands
	HostInfo := make(map[string]interface{}, 0)
	HostInfo, _ = util.ExecuteCommands(OSCommands, *dryrunPtr)
	HostInfo["Host"], _ = hostdiag.GetHost()
	HostInfo["Disk"], _ = hostdiag.GetDisk()
	HostInfo["Memory"], _ = hostdiag.GetMemory()
	HostInfo["Processes"], _ = hostdiag.GetProcesses()
	HostInfo["Network"], _ = hostdiag.GetNetwork()

	// Host info map into JSON
	HostInfoJSON, _ := util.MapToJSON(HostInfo)

	// Dump host info JSON into a results file
	util.JSONToFile(HostInfoJSON, dir+"/HostInfo.json")
	appLogger.Debug("Created output file", "name", hclog.Fmt("./%s", dir+"/HostInfo.json"))

	// Optional product commands
	if *productPtr != "" {
		appLogger.Info("Gathering product diagnostics")
		// Get Product Commands to run along with their attribute identifier, arguments, and format
		ProductCommands := make([]util.CommandStruct, 0)
		ProductCommands = products.ProductCommands(*productPtr)

		// Create map for product info and execute product commands
		ProductInfo := make(map[string]interface{}, 0)
		ProductInfo, _ = util.ExecuteCommands(ProductCommands, *dryrunPtr)

		// Product info map into JSON
		ProductInfoJSON, _ := util.MapToJSON(ProductInfo)

		// Dump product info JSON into a results_product file
		util.JSONToFile(ProductInfoJSON, dir+"/ProductInfo.json")
		appLogger.Debug("Created output file", "name", hclog.Fmt("./%s", dir+"/ProductInfo.json"))
	}

	// Create and compress archive of files in temporary folder
	appLogger.Info("Compressing and archiving output file", "name", *outfilePtr)
	util.TarGz(dir, "./"+*outfilePtr)
}
