package hostdiag

import (
	"fmt"
	"net"
	"runtime"

	"github.com/hashicorp/go-hclog"
	s "github.com/hashicorp/host-diagnostics/seeker"
	"github.com/hashicorp/host-diagnostics/util"
	"github.com/mitchellh/go-ps"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

func NewHostSeeker(os string) *s.Seeker {
	if os == "auto" {
		os = runtime.GOOS
	}
	return &s.Seeker{
		Identifier: "host",
		Runner: &HostSeeker{
			OS: os,
		},
	}
}

type HostSeeker struct {
	OS string `json:"os"`
}

func (hs *HostSeeker) Run() (interface{}, error) {
	results := make(map[string]interface{})

	// TODO: not throw away errors?  separate Seekers for each?  sheesh.
	results["uname"], _ = s.NewCommander("uname -v", "string", false).Run()
	results["host"], _ = GetHost()
	results["memory"], _ = GetMemory()
	results["disk"], _ = GetDisk()
	// TODO: change and/or uncomment these noisy things
	// results["processes"], _ = GetProcesses()
	// results["network"], _ = GetNetwork()

	return results, nil
}

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
