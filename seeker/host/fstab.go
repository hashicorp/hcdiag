package host

import (
	"fmt"

	"github.com/hashicorp/hcdiag/seeker"
)

var _ seeker.Runner = FSTab{}

type FSTab struct {
	os      string
	sheller *seeker.Seeker
}

func NewFSTab(os string) *seeker.Seeker {
	return &seeker.Seeker{
		Identifier: "/etc/fstab",
		Runner: FSTab{
			os:      os,
			sheller: seeker.NewSheller("cat /etc/fstab"),
		},
	}
}

func (s FSTab) Run() (interface{}, seeker.Status, error) {
	// Only Linux is supported currently; Windows is unsupported, and Darwin doesn't use /etc/fstab by default.
	if s.os != "linux" {
		// TODO(nwchandler): This should be seeker.Status("skip") once we implement it
		return nil, seeker.Success, fmt.Errorf("FSTab.Run() not available on os, os=%s", s.os)
	}
	res, _, err := s.sheller.Runner.Run()
	if err != nil {
		return res, seeker.Fail, err
	}

	return res, seeker.Success, nil
}
