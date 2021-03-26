package hostdiag

import (
	"fmt"
	"net"
	"runtime"

	"github.com/mitchellh/go-ps"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// OSCommand struct
type OSCommand struct {
	Attribute string
	Command   string
	Arguments []string
}

// OSCommands stuff
func OSCommands(operatingSystem string) []OSCommand {
	OSCommands := make([]OSCommand, 0)

	if operatingSystem == "auto" {
		operatingSystem = runtime.GOOS
	}

	switch {
	case operatingSystem == "darwin":
		OSCommands = append(OSCommands,
			OSCommand{
				Attribute: "Kernel",
				Command:   "uname",
				Arguments: []string{"-s"},
			},
			OSCommand{
				Attribute: "Kernel Release",
				Command:   "uname",
				Arguments: []string{"-r"},
			},
			OSCommand{
				Attribute: "Kernel Version",
				Command:   "uname",
				Arguments: []string{"-v"},
			})

	case operatingSystem == "linux":
		OSCommands = append(OSCommands,
			OSCommand{
				Attribute: "Kernel",
				Command:   "uname",
				Arguments: []string{"-s"},
			},
			OSCommand{
				Attribute: "Kernel Release",
				Command:   "uname",
				Arguments: []string{"-r"},
			},
			OSCommand{
				Attribute: "Kernel Version",
				Command:   "uname",
				Arguments: []string{"-v"},
			})

	default:
		fmt.Println("other os")

	}

	OSCommands = append(OSCommands,
		OSCommand{
			Attribute: "pwd",
			Command:   "pwd",
			Arguments: nil,
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
