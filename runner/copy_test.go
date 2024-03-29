// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package runner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/hcdiag/redact"

	"github.com/stretchr/testify/assert"
)

func TestNewCopy(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	now := time.Now()

	tt := []struct {
		desc      string
		cfg       CopyConfig
		expect    *Copy
		expectErr bool
	}{
		{
			desc: "valid config",
			cfg: CopyConfig{
				Path:    src,
				DestDir: dst,
				Since:   time.Time{},
				Until:   now,
			},
			expect: &Copy{
				ctx:       context.Background(),
				SourceDir: src,
				Filter:    "*",
				DestDir:   dst,
				Since:     time.Time{},
				Until:     now,
			},
		},
		{
			desc:      "empty config causes an error",
			cfg:       CopyConfig{},
			expectErr: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			c, err := NewCopy(tc.cfg)
			if tc.expectErr {
				assert.ErrorAs(t, err, &CopyConfigError{})
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expect, c)
			}
		})
	}
}

func setupFile(t *testing.T, dir, file, content string) {
	// touch file
	absFile := filepath.Join(dir, file)
	f, err := os.Create(absFile)
	assert.NoError(t, err)
	defer f.Close()

	// write to file
	_, err = f.WriteString(content)
	assert.NoError(t, err)
}

// TestFiles maps any number of filenames to their desired content
type TestFiles map[string]string

func TestCopyDir(t *testing.T) {
	tcs := []struct {
		name    string
		files   TestFiles
		redacts func(*testing.T) []*redact.Redact
	}{
		{
			name:  "can copy dir with empty file",
			files: TestFiles{"filename1": ""},
		},
		{
			name: "can copy dir with several empty files",
			files: TestFiles{
				"filename1": "",
				"file2.txt": "",
			},
		},
		{
			name:  "can copy single-file directory",
			files: TestFiles{"filename1": "Some content here"},
		},
		{
			name: "can copy multi-file directory",
			files: TestFiles{
				"filename1": "Some content here",
				"file2.txt": "more file content",
			},
		},
		{
			name: "can copy mixed multi-file directory that includes an empty file",
			files: TestFiles{
				"filename1": "Some content here",
				"file2.txt": "more file content",
				"empty":     "",
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var reds []*redact.Redact
			srcDir := t.TempDir()
			destDir := t.TempDir()

			if tc.redacts != nil {
				reds = tc.redacts(t)
			}

			// Create testfiles and content
			for name, content := range tc.files {
				setupFile(t, srcDir, name, content)
			}

			// Copy the directory contents
			ce := copyDir(destDir, srcDir, reds)
			assert.NoError(t, ce, tc.name)

			// Compare destination testfiles content
			for name, content := range tc.files {
				if name != "" {
					// confirm the file exists in the right place and has the right contents
					expectedLocation := filepath.Join(srcDir, name)
					result, err := os.ReadFile(expectedLocation)
					assert.NoError(t, err, tc.name)
					assert.Equal(t, content, string(result), tc.name)
				}
			}

		})
	}
}

func TestCopyDirErrors(t *testing.T) {
	tcs := []struct {
		name string
		dest string
		src  string
	}{
		{
			name: "empty src and dest",
		},
		{
			name: "dir does not exist",
			dest: "/tmp/faketestdestination",
			src:  "dir-doesnt-exist2347890-12348079-",
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := copyDir(tc.dest, tc.src, nil)
			assert.Error(t, err, tc.name)
		})
	}
}

func TestCopyFile(t *testing.T) {
	tcs := []struct {
		name  string
		files TestFiles
	}{
		{
			name:  "Test copying a single file",
			files: TestFiles{"filename1": "hack the planet"},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var reds []*redact.Redact
			srcDir := t.TempDir()
			dstDir := t.TempDir()

			for name, content := range tc.files {
				setupFile(t, srcDir, name, content)
				srcLocation := filepath.Join(srcDir, name)
				dstLocation := filepath.Join(dstDir, name)

				// Copy the file
				err := copyFile(dstLocation, srcLocation, reds)
				assert.NoError(t, err, tc.name)

				// Read the file
				result, err := os.ReadFile(dstLocation)
				assert.NoError(t, err, tc.name)
				assert.Equal(t, content, string(result), tc.name)
			}
		})
	}
}

// TODO(mkcp): more cases
func TestCopyFileErrors(t *testing.T) {
	tcs := []struct {
		name       string
		dest       string
		src        string
		redactions func(*testing.T) []*redact.Redact
	}{
		{
			name: "empty src and dest",
		},
		{
			name: "file does not exist",
			dest: "/tmp/faketestdestination",
			src:  "file-doesnt-exist1234712347091234",
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var reds []*redact.Redact
			if tc.redactions != nil {
				reds = tc.redactions(t)
			}
			err := copyFile(tc.dest, tc.src, reds)
			assert.Error(t, err, tc.name)
		})
	}
}
