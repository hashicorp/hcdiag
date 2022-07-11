package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/util"
)

var _ Runner = Copier{}

// Copier copies files to temp dir based on a filter.
type Copier struct {
	SourceDir string    `json:"source_directory"`
	Filter    string    `json:"filter"`
	DestDir   string    `json:"destination_directory"`
	Since     time.Time `json:"since"`
	Until     time.Time `json:"until"`
}

// NewCopier provides a Runner for copying files to temp dir based on a filter.
func NewCopier(path, destDir string, since, until time.Time) *Copier {
	sourceDir, filter := util.SplitFilepath(path)
	return &Copier{
		SourceDir: sourceDir,
		Filter:    filter,
		DestDir:   destDir,
		Since:     since,
		Until:     until,
	}
}

func (c Copier) ID() string {
	return "copy " + filepath.Join(c.SourceDir, c.Filter)
}

// Run satisfies the Runner interface and copies the filtered source files to the destination.
func (c Copier) Run() op.Op {
	// Ensure destination directory exists
	err := os.MkdirAll(c.DestDir, 0755)
	if err != nil {
		return op.New(c.ID(), nil, op.Fail, MakeDirError{
			path: c.DestDir,
			err:  err,
		},
			Params(c))
	}

	// Find all the files
	files, err := util.FilterWalk(c.SourceDir, c.Filter, c.Since, c.Until)
	if err != nil {
		return op.New(c.ID(), nil, op.Fail, FindFilesError{
			path: c.SourceDir,
			err:  err,
		},
			Params(c))
	}

	// Copy the files
	for _, s := range files {
		err = util.CopyDir(c.DestDir, s)
		if err != nil {
			return op.New(c.ID(), nil, op.Fail, CopyFilesError{
				dest:  c.DestDir,
				files: files,
				err:   err,
			},
				Params(c))
		}
	}

	return op.New(c.ID(), files, op.Success, nil, Params(c))
}

type MakeDirError struct {
	path string
	err  error
}

func (e MakeDirError) Error() string {
	return fmt.Sprintf("unable to mkdir, path=%s, err=%s", e.path, e.err.Error())
}

func (e MakeDirError) Unwrap() error {
	return e.err
}

type FindFilesError struct {
	path string
	err  error
}

func (e FindFilesError) Error() string {
	return fmt.Sprintf("unable to find files, path=%s, err=%s", e.path, e.err.Error())
}

func (e FindFilesError) Unwrap() error {
	return e.err
}

type CopyFilesError struct {
	dest  string
	files []string
	err   error
}

func (e CopyFilesError) Error() string {
	return fmt.Sprintf("unable to copy files, dest=%s, files=%v, err=%s", e.dest, e.files, e.err.Error())
}

func (e CopyFilesError) Unwrap() error {
	return e.err
}
