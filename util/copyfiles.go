package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-hclog"
)

const directoryPerms = 0755

// CopyDir copies a directory and all of its contents into a target directory.
func CopyDir(to, src string) error {
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
			if err := os.MkdirAll(target, directoryPerms); err != nil {
				return fmt.Errorf("os.MkdirAll failed: %w", err)
			}
		}
		if err := CopyFile(target, path); err != nil {
			return fmt.Errorf("CopyFile failed: %w", err)
		}
		return nil
	})
}

// CopyFile copies a file to a target file path.
func CopyFile(to, src string) error {
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

	// Write source contents to destination
	_, err = io.Copy(w, r)
	return err
}
