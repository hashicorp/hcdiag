package seeker

import (
	"github.com/hashicorp/host-diagnostics/util"
)

// NewCopier provides a Seeker for copying files to temp dir based on a filter.
func NewCopier(sourceDir string, filter string, destDir string, mustSucceed bool) *Seeker {
	return &Seeker{
		Identifier:  "copy " + sourceDir + "/" + filter,
		Runner:      Copier{SourceDir: sourceDir, Filter: filter, DestDir: destDir},
		MustSucceed: mustSucceed,
	}
}

// Copier copies files to temp dir based on a filter.
type Copier struct {
	SourceDir string `json:"source_directory"`
	Filter    string `json:"filter"`
	DestDir   string `json:"destination_directory"`
}

// Run satisfies the Runner interface and copies the filtered source files to the destination.
func (c Copier) Run() (result interface{}, err error) {
	files, err := util.FilterWalk(c.SourceDir, c.Filter)
	if err != nil {
		return nil, err
	}

	// Copy the files
	for _, s := range files {
		err := util.CopyDir(c.DestDir, s)
		if err != nil {
			return nil, err
		}
	}

	return files, nil
}
