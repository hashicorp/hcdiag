package util

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcdiag/redact"

	"github.com/hashicorp/go-hclog"
)

const directoryPerms = 0755

// CopyDir copies a directory and all of its contents into a target directory.
func CopyDir(to, src string, redactions []*redact.Redact) error {
	// get the absolute path, so we can remove it
	// to avoid copying the entire directory structure into the dest
	absPath, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for '%s': %s", src, err)
	}
	if _, err := os.Stat(absPath); err != nil {
		return fmt.Errorf("Expect %s to exist, got error: %s", absPath, err)
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
		return CopyFile(target, path, redactions)
	})
}

// CopyFile copies a file to a target file path.
func CopyFile(to, src string, redactions []*redact.Redact) error {
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
