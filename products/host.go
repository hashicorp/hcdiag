package products

import (
	"net"
	"runtime"

	"github.com/hashicorp/go-hclog"
	s "github.com/hashicorp/hcdiag/seeker"
	"github.com/mitchellh/go-ps"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// NewHost takes a product config and creates a Product containing all of the host's seekers.
func NewHost(cfg Config) *Product {
	return &Product{
		Seekers: []*s.Seeker{
			HostSeekers(cfg.OS),
		},
	}
}

func HostSeekers(os string) *s.Seeker {
	if os == "auto" {
		os = runtime.GOOS
	}
	return &s.Seeker{
		Identifier: "stats",
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
	results["uname"], _ = s.NewCommander("uname -v", "string").Run()
	results["host"], _ = GetHost()
	results["memory"], _ = GetMemory()
	results["disk"], _ = GetDisk()
	// TODO: change and/or uncomment these noisy things
	// results["processes"], _ = GetProcesses()
	// results["network"], _ = GetNetwork()

	return results, nil
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
