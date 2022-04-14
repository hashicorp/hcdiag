package seeker

import (
	"net"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/go-ps"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

type Host struct {
	OS string `json:"os"`
}

func (hs *Host) Run() (interface{}, Status, error) {
	results := make(map[string]interface{})
	var errors *multierror.Error

	if tmpResult, err := GetHost(); err != nil {
		errors = multierror.Append(errors, err)
	} else {
		results["host"] = tmpResult
	}
	if tmpResult, err := GetMemory(); err != nil {
		errors = multierror.Append(errors, err)
	} else {
		results["memory"] = tmpResult
	}
	if tmpResult, err := GetProcesses(); err != nil {
		errors = multierror.Append(errors, err)
	} else {
		results["processes"] = tmpResult
	}
	if tmpResult, err := GetNetwork(); err != nil {
		errors = multierror.Append(errors, err)
	} else {
		results["network"] = tmpResult
	}

	return results, Success, errors.ErrorOrNil()
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
