package seeker

import (
	"fmt"
	"net"
	"runtime"

	"github.com/mitchellh/go-ps"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// OSInfoCommand returns a command that can be run to gather information about the operating system
func OSInfoCommand() string {
	if runtime.GOOS == "windows" {
		return "systeminfo"
	}
	return "uname -v"
}

// HostInfo is a GoFunc that gathers information about the host machine
func HostInfo() (interface{}, Status, error) {
	info, err := host.Info()
	if err != nil {
		return info, Fail, fmt.Errorf("error getting host information err=%w", err)
	}
	return info, Success, nil
}

// DiskPartitions is a GoFunc that gathers information about disks
func DiskPartitions(all bool) func() (interface{}, Status, error) {
	return func() (interface{}, Status, error) {
		disks, err := disk.Partitions(all)
		if err != nil {
			return disks, Fail, fmt.Errorf("error getting disk information err=%w", err)
		}
		return disks, Success, nil
	}
}

// HostInfo is a GoFunc that gathers information about memory
func Memory() (interface{}, Status, error) {
	mem, err := mem.VirtualMemory()
	if err != nil {
		return mem, Fail, fmt.Errorf("error getting memory information err=%w", err)
	}
	return mem, Success, nil
}

// HostInfo is a GoFunc that gathers information about networks
func NetInterfaces() (interface{}, Status, error) {
	nets, err := net.Interfaces()
	if err != nil {
		return nets, Fail, fmt.Errorf("error getting network information err=%w", err)
	}
	return nets, Success, nil
}

// GetProcesses is a GoFunc that gets a list of running process names
func GetProcesses() (interface{}, Status, error) {
	processes, err := ps.Processes()
	if err != nil {
		return processes, Fail, fmt.Errorf("error getting process information err=%w", err)
	}

	processInfo := make([]string, 0)

	for eachProcess := range processes {
		process := processes[eachProcess]
		processInfo = append(processInfo, process.Executable())
	}

	return processInfo, Success, nil
}
