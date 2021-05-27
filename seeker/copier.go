package seeker

import (
	"os"
	"time"

	"github.com/hashicorp/host-diagnostics/util"
)

// NewCopier provides a Seeker for copying files to temp dir based on a filter.
func NewCopier(path, destDir string, from, to time.Time) *Seeker {
	sourceDir, filter := util.SplitFilepath(path)
	return &Seeker{
		Identifier: "copy " + sourceDir + "/" + filter,
		Runner: Copier{
			SourceDir: sourceDir,
			Filter: filter,
			DestDir: destDir,
			From: from,
			To: to,
		},
	}
}

// Copier copies files to temp dir based on a filter.
type Copier struct {
	SourceDir string    `json:"source_directory"`
	Filter    string    `json:"filter"`
	DestDir   string    `json:"destination_directory"`
	From      time.Time `json:"from"`
	To        time.Time `json:"to"`
}

// Run satisfies the Runner interface and copies the filtered source files to the destination.
func (c Copier) Run() (result interface{}, err error) {
	// Ensure destination directory exists
	err = os.MkdirAll(c.DestDir, 0755)
	if err != nil {
		return nil, err
	}

	// Find all the files
	files, err := util.FilterWalk(c.SourceDir, c.Filter, c.From, c.To)
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
