package hostdiag

import (
	"fmt"
	"net"
	"runtime"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/host-diagnostics/util"
	"github.com/mitchellh/go-ps"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// OSCommands stuff
func OSCommands(operatingSystem string) []util.CommandStruct {
	OSCommands := make([]util.CommandStruct, 0)

	if operatingSystem == "auto" {
		operatingSystem = runtime.GOOS
	}

	switch {
	case operatingSystem == "darwin":
		OSCommands = append(OSCommands,
			util.CommandStruct{
				Attribute: "Kernel",
				Command:   "uname",
				Arguments: []string{"-s"},
				Format:    "string",
			},
			util.CommandStruct{
				Attribute: "Kernel Release",
				Command:   "uname",
				Arguments: []string{"-r"},
				Format:    "string",
			},
			util.CommandStruct{
				Attribute: "Kernel Version",
				Command:   "uname",
				Arguments: []string{"-v"},
				Format:    "string",
			})

	case operatingSystem == "linux":
		OSCommands = append(OSCommands,
			util.CommandStruct{
				Attribute: "Kernel",
				Command:   "uname",
				Arguments: []string{"-s"},
				Format:    "string",
			},
			util.CommandStruct{
				Attribute: "Kernel Release",
				Command:   "uname",
				Arguments: []string{"-r"},
				Format:    "string",
			},
			util.CommandStruct{
				Attribute: "Kernel Version",
				Command:   "uname",
				Arguments: []string{"-v"},
				Format:    "string",
			})

	default:
		fmt.Println("other os")

	}

	OSCommands = append(OSCommands,
		util.CommandStruct{
			Attribute: "pwd",
			Command:   "pwd",
			Arguments: nil,
			Format:    "string",
		})

	return OSCommands
}

// GetNetwork stuff
func GetNetwork() (interface{}, error) {
	networkInfo, err := net.Interfaces()
	if err != nil {
		hclog.L().Error("GetNetwork", "Error getting network information", err)
		return networkInfo, err
	}

	return networkInfo, err
}

// GetProcesses stuff
func GetProcesses() (interface{}, error) {
	processes, err := ps.Processes()
	if err != nil {
		hclog.L().Error("GetProcesses", "Error getting process information", err)
		return processes, err
	}

	processInfo := make([]string, 0)

	for eachProcess := range processes {
		var process ps.Process
		process = processes[eachProcess]
		processInfo = append(processInfo, process.Executable())
	}

	return processInfo, err
}

// basic functions below serving mostly as placeholders for third party libs
// -------------------------------------------------------------------------

// GetMemory stuff
func GetMemory() (interface{}, error) {
	// third party
	memoryInfo, err := mem.VirtualMemory()
	if err != nil {
		hclog.L().Error("GetMemory", "Error getting memory information", err)
		return memoryInfo, err
	}

	return memoryInfo, err
}

// GetDisk stuff
func GetDisk() (interface{}, error) {
	// third party
	diskInfo, err := disk.Partitions(true)
	if err != nil {
		hclog.L().Error("GetDisk", "Error getting disk information", err)
		return diskInfo, err
	}

	return diskInfo, err
}

// GetHost stuff
func GetHost() (interface{}, error) {
	// third party
	hostInfo, err := host.Info()
	if err != nil {
		hclog.L().Error("GetHost", "Error getting host information", err)
		return hostInfo, err
	}

	return hostInfo, err
}
