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

	hs.capture("host", func() (interface{}, error) { return host.Info() })
	hs.capture("memory", func() (interface{}, error) { return mem.VirtualMemory() })
	hs.capture("disk", func() (interface{}, error) { return disk.Partitions(true) })
	hs.capture("processes", func() (interface{}, error) { return ps.Processes() })
	hs.capture("network", func() (interface{}, error) { return net.Interfaces() })

	l.Debug("Host seeker capture complete")
	return hs.results, hs.errors.ErrorOrNil()
}
