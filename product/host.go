package product

import (
	"errors"
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
			s.NewCommander(OSInfoCommand(), "string"),
			s.NewGoFuncSeeker("host", func() (interface{}, error) { return host.Info() }),
			s.NewGoFuncSeeker("disks", func() (interface{}, error) { return disk.Partitions(true) }),
			s.NewGoFuncSeeker("memory", func() (interface{}, error) { return mem.VirtualMemory() }),
			s.NewGoFuncSeeker("network", func() (interface{}, error) { return net.Interfaces() }),
			// i have heard that this is not super useful, since it's just process names...
			s.NewGoFuncSeeker("processes", GetProcesses),

			// example failure case
			s.NewGoFuncSeeker("bad-and-not-good", func() (interface{}, error) { return nil, errors.New("no good at all") }),
		},
	}
}

// OSInfoCommand returns a command that can be run to gather information about the operating system
func OSInfoCommand() string {
	if runtime.GOOS == "windows" {
		return "systeminfo"
	}
	return "uname -v"
}

// GetProcesses gets a list of executable names of running processes
func GetProcesses() (interface{}, error) {
	processes, err := ps.Processes()
	if err != nil {
		hclog.L().Error("GetProcesses", "Error getting process information", err)
		return processes, err
	}

	processInfo := make([]string, 0)

	for eachProcess := range processes {
		process := processes[eachProcess]
		processInfo = append(processInfo, process.Executable())
	}

	return processInfo, err
}
