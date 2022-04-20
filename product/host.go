package product

import (
	"errors"

	s "github.com/hashicorp/hcdiag/seeker"
)

// NewHost takes a product config and creates a Product containing all of the host's seekers.
func NewHost(cfg Config) *Product {
	return &Product{
		Seekers: []*s.Seeker{
			s.NewCommander(s.OSInfoCommand(), "string"),

			s.NewGoFuncSeeker("host", s.HostInfo),
			s.NewGoFuncSeeker("disks", s.DiskPartitions(false)),
			s.NewGoFuncSeeker("memory", s.Memory),
			s.NewGoFuncSeeker("network", s.NetInterfaces),
			// i have heard that this is not super useful, since it's just process names...
			s.NewGoFuncSeeker("processes", s.GetProcesses),

			// example failure case
			s.NewGoFuncSeeker("bad-and-not-good", BadAndNotGood),
		},
	}
}

func BadAndNotGood() (interface{}, s.Status, error) {
	return nil, s.Fail, errors.New("no good at all")
}
