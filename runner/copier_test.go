package runner

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/hcdiag/redact"

	"github.com/stretchr/testify/assert"
)

func TestNewCopier(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()
	since := time.Time{}
	until := time.Now()
	expect := &Copier{
		SourceDir: src,
		Filter:    "*",
		DestDir:   dest,
		Since:     since,
		Until:     until,
	}
	copier := NewCopier(src, dest, since, until, nil)
	assert.Equal(t, expect, copier)
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

// filename -> content
type TestFileList map[string]string

func TestCopyDir(t *testing.T) {
	tcs := []struct {
		name    string
		files   TestFileList
		redacts func(*testing.T) []*redact.Redact
	}{
		{
			name:  "can copy dir with empty file",
			files: TestFileList{"filename1": ""},
		},
		{
			name: "can copy dir with several empty files",
			files: TestFileList{
				"filename1": "",
				"file2.txt": "",
			},
		},
		{
			name:  "can copy single-file directory",
			files: TestFileList{"filename1": "Some content here"},
		},
		{
			name: "can copy multi-file directory",
			files: TestFileList{
				"filename1": "Some content here",
				"file2.txt": "more file content",
			},
		},
		{
			name: "can copy mixed multi-file directory that includes an empty file",
			files: TestFileList{
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
		err := copyDir(tc.dest, tc.src, nil)
		assert.Error(t, err, tc.name)
	}
}

// TODO(mkcp): This code is exercised by the CopyDir cases above, but it should be tested in isolation with a variety
//  of cases.
func TestCopyFile(t *testing.T) {
	t.Skip()
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
		var reds []*redact.Redact
		if tc.redactions != nil {
			reds = tc.redactions(t)
		}
		err := copyFile(tc.dest, tc.src, reds)
		assert.Error(t, err, tc.name)
	}
}
