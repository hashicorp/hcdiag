package main

import (
	"flag"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/host-diagnostics/hostdiag"
	"github.com/hashicorp/host-diagnostics/products"
	"github.com/hashicorp/host-diagnostics/util"
)

func main() {
	// TODO: eval third party libs, gap and risk analysis
	// TODO: update data model, lots of things generic currently
	// TODO: determine appropriate arguments, eval cli libs
	// TODO: standardize error and exception handling
	// TODO: add support to targz for multiple files / dir, improve func; found several good examples but wanted to understand myself before using
	// TODO: expand os and product cmds, os commands are really just placeholders atm
	// TODO: expand hostdiag process, currently only returning all process names and not very useful
	// TODO: add outfile arg logic or similar, possibly options for output type
	// TODO: consolidate output into single targz

	// Create logger
	appLogger := hclog.New(&hclog.LoggerOptions{
		Name: "host-diagnostics",
	})
	if logStr := os.Getenv("LOG_LEVEL"); logStr != "" {
		if level := hclog.LevelFromString(logStr); level != hclog.NoLevel {
			appLogger.SetLevel(level)
			appLogger.Trace("logger configuration change", "LOG_LEVEL", hclog.Fmt("%s", logStr))
		}
	}

	osPtr := flag.String("os", "auto", "(optional) override operating system detection")                // auto, darwin, linux, ??
	productPtr := flag.String("product", "", "(optional) run product diagnostic commands if specified") // terraform, vault, ??
	dryrunPtr := flag.Bool("dryrun", false, "(optional) performing a dry run will display all commands without executing them")
	// levelPtr := flag.String("level", "full", "(optional) info gathering level")     // basic, enhanced, full ??
	// outfilePtr := flag.String("outfile", "./outfile", "(optional) output filepath") // ./outfile, diff types??
	flag.Parse()

	// Get list of OS Commands to run along with their attribute identifier, arguments, and format
	OSCommands := make([]util.CommandStruct, 0)
	OSCommands = hostdiag.OSCommands(*osPtr)

	// Create map for host info and execute os commands
	DiagInfo := make(map[string]interface{}, 0)
	DiagInfo = util.ExecuteCommands(OSCommands, *dryrunPtr)
	DiagInfo["Host"] = hostdiag.GetHost()
	DiagInfo["Disk"] = hostdiag.GetDisk()
	DiagInfo["Memory"] = hostdiag.GetMemory()
	DiagInfo["Processes"] = hostdiag.GetProcesses()
	DiagInfo["Network"] = hostdiag.GetNetwork()

	// Optional product commands
	if *productPtr != "" {
		// Get Product Commands to run along with their attribute identifier, arguments, and format
		ProductCommands := make([]util.CommandStruct, 0)
		ProductCommands = products.ProductCommands(*productPtr)

		// Create map for product info and execute product commands
		ProductInfo := make(map[string]interface{}, 0)
		ProductInfo = util.ExecuteCommands(ProductCommands, *dryrunPtr)

		// Product info map into JSON
		ProductInfoJSON := util.MapToJSON(ProductInfo)

		// Dump product info JSON into a results_product file
		util.JSONToFile(ProductInfoJSON, "./results_product.json")
	}

	// TODO: consolidate output
	// ------------------------

	// Host info map into JSON
	DiagInfoJSON := util.MapToJSON(DiagInfo)

	// Dump host info JSON into a results file
	util.JSONToFile(DiagInfoJSON, "./results.json")

	// Create and compress archive
	util.TarGz("./results.json", "./results.tar", "./results.tar.gz")
}
