package seeker

import (
	"github.com/hashicorp/host-diagnostics/util"
)

// NewCopier provides a Seeker for copying files to temp dir based on a filter.
func NewCopier(SourceDir string, Filter string, DestDir string, mustSucceed bool) *Seeker {
	return &Seeker{
		Identifier:  "copy " + SourceDir + "/" + Filter,
		Runner:      Copier{SourceDir: SourceDir, Filter: Filter, DestDir: DestDir},
		MustSucceed: mustSucceed,
	}
}

// Copier copies files to temp dir based on a filter.
type Copier struct {
	SourceDir string `json:"source_directory"`
	Filter    string `json:"filter"`
	DestDir   string `json:"destination_directory"`
}

func (c Copier) Run() (result interface{}, err error) {
	files, err := util.FilterWalk(c.SourceDir, c.Filter)
	if err != nil {
		return nil, err
	}

	for _, s := range files {
		err := util.CopyDir(c.DestDir, s)
		if err != nil {
			return nil, err
		}
	}

	return files, nil
}
