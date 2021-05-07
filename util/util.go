package util

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
)

// TarGz accepts a source directory and destination file name to archive and compress files.
func TarGz(sourceDir string, destFileName string) error {
	// tar
	destFile, err := os.Create(destFileName)
	if err != nil {
		hclog.L().Error("TarGz", "error creating tarball", err)
		return err
	}
	defer destFile.Close()

	gzWriter := gzip.NewWriter(destFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() == false {
			sourceFile, err := os.Open(path)
			if err != nil {
				hclog.L().Error("TarGz", "error opening source file", err)
				return err
			}
			defer sourceFile.Close()

			stat, err := sourceFile.Stat()
			if err != nil {
				hclog.L().Error("TarGz", "error checking source file", err)
				return err
			}

			header := &tar.Header{
				Name:    path,
				Size:    stat.Size(),
				Mode:    int64(stat.Mode()),
				ModTime: stat.ModTime(),
			}

			if err := tarWriter.WriteHeader(header); err != nil {
				hclog.L().Error("TarGz", "error writing header for tar", err)
				return err
			}

			if _, err := io.Copy(tarWriter, sourceFile); err != nil {
				hclog.L().Error("TarGz", "error copying file to tarball", err)
				return err
			}
		}
		return nil
	})

	return err
}

// WriteJSON converts an interface{} to JSON then writes to filePath.
func WriteJSON(iface interface{}, filePath string) error {
	// TODO: these funcs have their own logging in em, perhaps we don't want that?
	jsonBts, err := InterfaceToJSON(iface)
	if err != nil {
		return err
	}
	err = JSONToFile(jsonBts, filePath)
	if err != nil {
		return err
	}
	return nil
}

// InterfaceToJSON converts an interface{} to JSON.
func InterfaceToJSON(mapVar interface{}) ([]byte, error) {
	InfoJSON, err := json.MarshalIndent(mapVar, "", "    ")
	if err != nil {
		hclog.L().Error("InterfaceToJSON", "message", err)
		return InfoJSON, err
	}

	return InfoJSON, err
}

// JSONToFile accepts JSON and an output file path to create a JSON file.
func JSONToFile(JSON []byte, outFile string) error {
	err := ioutil.WriteFile(outFile, JSON, 0644)
	if err != nil {
		hclog.L().Error("JSONToFile", "error during json to file", err)
	}

	return err
}

// FilterWalk accepts a source directory and filter to return a list of matching files.
func FilterWalk(srcDir, filter string) ([]string, error) {
	var fileMatches []string

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		// Check for files that match the filter
		match, err := filepath.Match(filter, filepath.Base(path))
		if match && err == nil {
			fileMatches = append(fileMatches, path)
		}

		return err
	})
	if err != nil {
		return nil, err
	}

	return fileMatches, nil
}
