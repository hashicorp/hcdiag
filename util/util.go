package util

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-hclog"
)

// CommandStruct stuff
type CommandStruct struct {
	Attribute string
	Command   string
	Arguments []string
	Format    string
}

// ExecuteCommands stuff
func ExecuteCommands(CommandList []CommandStruct, dryrunPtr bool) (map[string]interface{}, error) {
	// Create map for info
	ReturnInfo := make(map[string]interface{}, 0)

	// Run Commands
	for _, element := range CommandList {
		hclog.L().Debug("ExecuteCommands", "command", element.Command, "arguments", element.Arguments)

		if dryrunPtr == false {
			CommandOutput, err := exec.Command(element.Command, element.Arguments...).CombinedOutput()
			if err != nil {
				hclog.L().Error("ExecuteCommands", "error during command execution", err, "command", element.Command, "arguments", element.Arguments)
				return ReturnInfo, err
			}

			// If format is json then unmarshal to interface{}, otherwise out string
			if element.Format == "json" {
				var outInterface interface{}
				if err := json.Unmarshal(CommandOutput, &outInterface); err != nil {
					hclog.L().Error("ExecuteCommands", "error during unmarshal to json", err)
					return ReturnInfo, err
				}
				ReturnInfo[element.Attribute] = outInterface

			} else {
				outString := strings.TrimSuffix(string(CommandOutput), "\n")
				ReturnInfo[element.Attribute] = outString
			}
		}
	}

	return ReturnInfo, nil
}

// TarGz func to archive and compress
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

// InterfaceToJSON stuff
func InterfaceToJSON(mapVar interface{}) ([]byte, error) {
	InfoJSON, err := json.MarshalIndent(mapVar, "", "    ")
	if err != nil {
		hclog.L().Error("InterfaceToJSON", "error during map to json", err)
		return InfoJSON, err
	}

	return InfoJSON, err
}

// JSONToFile stuff
func JSONToFile(JSON []byte, outFile string) error {
	err := ioutil.WriteFile(outFile, JSON, 0644)
	if err != nil {
		hclog.L().Error("JSONToFile", "error during json to file", err)
	}

	return err
}
