package host

import (
	"fmt"
	"runtime"

	"github.com/hashicorp/hcdiag/seeker"
)

var _ seeker.Runner = ProcFile{}

type ProcFile struct {
	os       string
	commands []string
}

func NewProcFile() *seeker.Seeker {
	return &seeker.Seeker{
		Identifier: "/proc/ files",
		Runner: ProcFile{
			os: runtime.GOOS,
			commands: []string{
				"cat /proc/cpuinfo",
				"cat /proc/loadavg",
				"cat /proc/version",
				"cat /proc/vmstat",
			},
		},
	}
}

func (s ProcFile) Run() (interface{}, seeker.Status, error) {
	if s.os != "linux" {
		// TODO(mkcp): Replace status with seeker.Skip when we implement it
		return nil, seeker.Success, fmt.Errorf("os not linux, skipping, os=%s", s.os)
	}
	m := make(map[string]interface{})
	for _, c := range s.commands {
		res, _, err := seeker.NewSheller(c).Runner.Run()
		m[c] = res
		if err != nil {
			return m, seeker.Fail, err
		}
	}
	return m, seeker.Success, nil
}
