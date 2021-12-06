package product

import (
	"net"
	"runtime"

	"github.com/hashicorp/go-hclog"
	multierror "github.com/hashicorp/go-multierror"
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
			l:       hclog.L().Named("host"),
			OS:      os,
			results: make(map[string]interface{}),
			errors:  new(multierror.Error),
		},
	}
}

type HostSeeker struct {
	l       hclog.Logger
	OS      string `json:"os"`
	results map[string]interface{}
	errors  *multierror.Error
}

type captureFunc func() (interface{}, error)

func (hs *HostSeeker) capture(name string, f captureFunc) {
	l := hs.l.Named("capture").With("name", name)
	l.Debug("Capturing host info")

	result, err := f()
	if err != nil {
		l.Error("Error capturing host info", "err", err)
		hs.errors = multierror.Append(hs.errors, err)
	}
	hs.results[name] = result

	l.Trace("Host capture results", "result", result)
}

func (hs *HostSeeker) Run() (interface{}, error) {
	l := hs.l.Named("Run")
	l.Debug("Host seeker capture begin")

	hs.capture("uname", s.NewCommander("uname -v", "string").Run)
	hs.capture("host", GetHost)
	hs.capture("memory", GetMemory)
	hs.capture("disk", GetDisk)
	hs.capture("processes", GetProcesses)
	hs.capture("network", GetNetwork)

	l.Debug("Host seeker capture complete")
	return hs.results, hs.errors.ErrorOrNil()
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
		process := processes[eachProcess]
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
