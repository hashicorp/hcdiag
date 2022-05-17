package host

import (
	"fmt"
	"runtime"

	"github.com/hashicorp/hcdiag/seeker"
)

var _ seeker.Runner = EtcHosts{}

type EtcHosts struct {
	os string
}

func NewEtcHosts() *seeker.Seeker {
	return &seeker.Seeker{
		Identifier: "/etc/hosts",
		Runner:     EtcHosts{os: runtime.GOOS},
	}
}

func (s EtcHosts) Run() (interface{}, seeker.Status, error) {
	// Not compatible with windows
	if s.os == "windows" {
		// TODO(mkcp): This should be seeker.Status("skip") once we implement it
		return nil, seeker.Success, fmt.Errorf(" EtcHosts.Run() not available on os, os=%s", s.os)
	}
	res, _, err := seeker.NewSheller("cat /etc/hosts").Runner.Run()
	if err != nil {
		return res, seeker.Fail, err
	}
	return res, seeker.Success, nil
}
