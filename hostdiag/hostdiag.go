package hostdiag

import (
	"fmt"
	"net"
	"runtime"

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
func GetNetwork() interface{} {
	networkInfo, err := net.Interfaces()
	if err != nil {
		return "Unable to get network info"
	}

	return networkInfo
}

// GetProcesses stuff
func GetProcesses() interface{} {
	processes, err := ps.Processes()
	if err != nil {
		return "Unable to get process info"
	}

	processInfo := make([]string, 0)

	for eachProcess := range processes {
		var process ps.Process
		process = processes[eachProcess]
		processInfo = append(processInfo, process.Executable())
	}

	return processInfo
}

// basic functions below serving mostly as placeholders for third party libs
// -------------------------------------------------------------------------

// GetMemory stuff
func GetMemory() interface{} {
	// third party
	memoryInfo, err := mem.VirtualMemory()
	if err != nil {
		return "Unable to get memory info"
	}

	return memoryInfo
}

// GetDisk stuff
func GetDisk() interface{} {
	// third party
	diskInfo, err := disk.Partitions(true)
	if err != nil {
		return "Unable to get disk info"
	}

	return diskInfo
}

// GetHost stuff
func GetHost() interface{} {
	// third party
	hostInfo, err := host.Info()
	if err != nil {
		return "Unable to get host info"
	}

	return hostInfo
}
