package seeker

import (
	"github.com/hashicorp/host-diagnostics/util"
)

// NewCopier provides a Seeker for copying files to temp dir based on a filter.
func NewCopier(srcDir string, filter string, tmpDir string, mustSucceed bool) *Seeker {
	return &Seeker{
		Identifier:  "copy files matching: " + filter + " from directory: " + srcDir,
		Runner:      Copier{srcDir: srcDir, filter: filter, tmpDir: tmpDir},
		MustSucceed: mustSucceed,
	}
}

// Copier copies files to temp dir based on a filter.
type Copier struct {
	srcDir string `json:"srcDir"`
	filter string `json:"filter"`
	tmpDir string `json:"tmpDir"`
}

func (c Copier) Run() (result interface{}, err error) {
	files, err := util.FilterWalk(c.srcDir, c.filter)
	if err != nil {
		return nil, err
	}

	for _, s := range files {
		err := util.CopyDir(c.tmpDir, s)
		if err != nil {
			return nil, err
		}
	}

	return files, nil
}
