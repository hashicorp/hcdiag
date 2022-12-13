// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package runner

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/hcdiag/op"

	"github.com/hashicorp/hcdiag/util"

	"github.com/hashicorp/go-hclog"
)

var _ Runner = Copy{}

type CopyConfig struct {
	// Path is the file path to the directory or file to copy to the DestDir.
	Path    string
	DestDir string
	Since   time.Time
	Until   time.Time
	// Redactions includes any redactions to apply to the output of the runner.
	Redactions []*redact.Redact
	// Timeout specifies the amount of time that the runner should be allowed to execute before cancellation.
	Timeout time.Duration
}

// Copy copies files to temp dir based on a filter.
type Copy struct {
	ctx context.Context

	SourceDir string    `json:"source_directory"`
	Filter    string    `json:"filter"`
	DestDir   string    `json:"destination_directory"`
	Since     time.Time `json:"since"`
	Until     time.Time `json:"until"`
	// Redactions includes any redactions to apply to the output of the runner.
	Redactions []*redact.Redact `json:"redactions"`
	// Timeout specifies the amount of time that the runner should be allowed to execute before cancellation.
	Timeout Timeout `json:"timeout"`
}

// NewCopy provides a Runner for copying files to temp dir based on a filter.
func NewCopy(cfg CopyConfig) (*Copy, error) {
	return NewCopyWithContext(context.Background(), cfg)
}

func NewCopyWithContext(ctx context.Context, cfg CopyConfig) (*Copy, error) {
	path := cfg.Path
	if path == "" {
		return nil, CopyConfigError{
			config: cfg,
			err:    fmt.Errorf("path must not be empty when creating a Copy runner"),
		}
	}

	sourceDir, filter := util.SplitFilepath(path)
	return &Copy{
		ctx: ctx,

		SourceDir:  sourceDir,
		Filter:     filter,
		DestDir:    cfg.DestDir,
		Since:      cfg.Since,
		Until:      cfg.Until,
		Redactions: cfg.Redactions,
		Timeout:    Timeout(cfg.Timeout),
	}, nil
}

func (c Copy) ID() string {
	return "copy " + filepath.Join(c.SourceDir, c.Filter)
}

// Run satisfies the Runner interface and copies the filtered source files to the destination.
func (c Copy) Run() op.Op {
	startTime := time.Now()

	// Ensure destination directory exists
	err := os.MkdirAll(c.DestDir, 0755)
	if err != nil {
		return op.New(c.ID(), nil, op.Fail,
			MakeDirError{
				path: c.DestDir,
				err:  err,
			}, Params(c), startTime, time.Now())
	}

	// Find all the files
	files, err := filterWalk(c.SourceDir, c.Filter, c.Since, c.Until)
	if err != nil {
		return op.New(c.ID(), nil, op.Fail,
			FindFilesError{
				path: c.SourceDir,
				err:  err,
			}, Params(c), startTime, time.Now())
	}

	// Copy the files
	for _, s := range files {
		err = copyDir(c.DestDir, s, c.Redactions)
		if err != nil {
			return op.New(c.ID(), nil, op.Fail,
				CopyFilesError{
					dest:  c.DestDir,
					files: files,
					err:   err,
				}, Params(c), startTime, time.Now())
		}
	}

	result := map[string]any{"files": files}
	return op.New(c.ID(), result, op.Success, nil, Params(c), startTime, time.Now())
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

// filterWalk accepts a source directory, filter string, and since and to Times to return a list of matching files.
func filterWalk(srcDir, filter string, since, until time.Time) ([]string, error) {
	var fileMatches []string

	// Filter the files
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		// Check for files that match the filter then check for time matches
		match, err := filepath.Match(filter, filepath.Base(path))
		if match && err == nil {
			// grab our file's last modified time
			info, err := os.Stat(path)
			if err != nil {
				return err
			}
			mod := info.ModTime()
			if util.IsInRange(mod, since, until) {
				fileMatches = append(fileMatches, path)
			}
		}
		return err
	})
	if err != nil {
		return nil, err
	}

	return fileMatches, nil
}

const directoryPerms = 0755

// copyDir copies a directory and all of its contents into a target directory.
func copyDir(to, src string, redactions []*redact.Redact) error {
	if src == "" {
		return fmt.Errorf("no source directory given, src=%s, to=%s", src, to)
	}
	// get the absolute path, so we can remove it
	// to avoid copying the entire directory structure into the dest
	absPath, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for '%s': %s", src, err)
	}
	if _, err := os.Stat(absPath); err != nil {
		return fmt.Errorf("expect %s to exist, got error: %s", absPath, err)
	}
	absBase := filepath.Dir(absPath)

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		// Windows path may contain unsafe characters
		targetMaybeUnsafe := filepath.Join(to, absBase, info.Name())

		// TODO: more extensive path cleansing beyond handling C:\
		target := strings.Replace(targetMaybeUnsafe, ":", "_", -1)

		if info.IsDir() {
			hclog.L().Info("copying", "path", path, "to", target)
			return os.MkdirAll(target, directoryPerms)
		}
		return copyFile(target, path, redactions)
	})
}

// copyFile copies a file to a target file path.
func copyFile(to, src string, redactions []*redact.Redact) error {
	hclog.L().Info("copying", "path", src, "to", to)

	// Ensure directories
	dir, _ := filepath.Split(to)
	err := os.MkdirAll(dir, directoryPerms)
	if err != nil {
		return err
	}

	// Open source file
	r, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			hclog.L().Error("Unable to close source file", "error", err)
		}
	}()

	// Create destination file
	w, err := os.Create(to)
	if err != nil {
		return err
	}
	defer func() {
		if err := w.Close(); err != nil {
			hclog.L().Error("Unable to close dest file", "error", err)
		}
	}()

	if 0 < len(redactions) {
		scanner := bufio.NewScanner(r)
		// Scan, redact, and write each line of the src file
		for scanner.Scan() {
			bts := scanner.Bytes()
			bts = append(bts, '\n')
			rBts, re := redact.Bytes(bts, redactions)
			if re != nil {
				return re
			}
			_, we := w.Write(rBts)
			if we != nil {
				return we
			}
		}
		return nil
	}

	// No redactions, copy as normal
	_, ce := io.Copy(w, r)
	return ce
}

var _ error = CommandConfigError{}

type CopyConfigError struct {
	config CopyConfig
	err    error
}

func (e CopyConfigError) Error() string {
	message := "invalid Copy Config"
	if e.err != nil {
		return fmt.Sprintf("%s: %s", message, e.err.Error())
	}
	return message
}

func (e CopyConfigError) Unwrap() error {
	return e.err
}
