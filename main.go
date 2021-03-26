package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os/exec"
	"strings"

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

	osPtr := flag.String("os", "auto", "(optional) operating system override") // auto, darwin, linux, ??
	productPtr := flag.String("product", "", "(optional) product name")        // terraform, vault, ??
	dryrunPtr := flag.Bool("dryrun", false, "(optional) performing a dry run will display all commands without executing them")
	// levelPtr := flag.String("level", "full", "(optional) info gathering level")     // basic, enhanced, full ??
	// outfilePtr := flag.String("outfile", "./outfile", "(optional) output filepath") // ./outfile, diff types??
	flag.Parse()

	hostInfo := hostdiag.GetHost()
	diskInfo := hostdiag.GetDisk()
	memoryInfo := hostdiag.GetMemory()
	processInfo := hostdiag.GetProcesses()
	networkInfo := hostdiag.GetNetwork()

	// Create map for host info
	DiagInfo := map[string]interface{}{
		"Host":      hostInfo,
		"Disk":      diskInfo,
		"Memory":    memoryInfo,
		"Processes": processInfo,
		"Network":   networkInfo,
	}

	// Get list of OS commands to execute
	OSCommands := make([]hostdiag.OSCommand, 0)
	OSCommands = hostdiag.OSCommands(*osPtr)

	// Run OS Commands and add to DiagInfo map
	for _, element := range OSCommands {
		fmt.Printf("%s %v\n", element.Command, element.Arguments)
		if *dryrunPtr == false {
			CommandOutput, err := exec.Command(element.Command, element.Arguments...).Output()
			if err != nil {
				fmt.Println(err)
			}
			DiagInfo[element.Attribute] = strings.TrimSuffix(string(CommandOutput), "\n")
		}
	}

	// Host info map into JSON
	DiagInfoJSON := util.MapToJSON(DiagInfo)

	// Dump host info JSON into a results file
	util.JSONToFile(DiagInfoJSON, "./results.json")

	// Create and compress archive
	util.TarGz("./results.json", "./results.tar", "./results.tar.gz")

	// Optional product commands
	if *productPtr != "" {
		// Create map for product info
		ProductInfo := map[string]interface{}{
			"Product": *productPtr,
		}

		// Get Product Commands to run along with their attribute identifier and arguments
		ProductCommands := make([]products.ProductCommand, 0)
		ProductCommands = products.ProductCommands(*productPtr)

		// Run Product Commands
		for _, element := range ProductCommands {
			fmt.Printf("%s %v\n", element.Command, element.Arguments)
			if *dryrunPtr == false {
				CommandOutput, err := exec.Command(element.Command, element.Arguments...).CombinedOutput()
				if err != nil {
					fmt.Println(err)
				}

				// If format is json then unmarshal to map[string]interface{}, otherwise out string
				if element.Format == "json" {
					var outInterface map[string]interface{}
					if err := json.Unmarshal(CommandOutput, &outInterface); err != nil {
						fmt.Println(err)
					}
					ProductInfo[element.Attribute] = outInterface
				} else {
					outString := strings.TrimSuffix(string(CommandOutput), "\n")
					ProductInfo[element.Attribute] = outString
				}
			}
		}

		// Product info map into JSON
		ProductInfoJSON := util.MapToJSON(ProductInfo)

		// Dump product info JSON into a results_product file
		util.JSONToFile(ProductInfoJSON, "./results_product.json")
	}
}
