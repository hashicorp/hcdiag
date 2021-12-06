package client

// https://www.nomadproject.io/api-docs

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const DefaultNomadAddr = "http://127.0.0.1:4646"

// NewNomadAPI returns an APIClient for Nomad.
func NewNomadAPI() *APIClient {
	addr := os.Getenv("NOMAD_ADDR")
	if addr == "" {
		addr = DefaultNomadAddr
	}

	headers := map[string]string{}
	token := os.Getenv("NOMAD_TOKEN")
	if token != "" {
		headers["X-Nomad-Token"] = token
	}

	return NewAPIClient("nomad", addr, headers)
}

func GetNomadLogPath(api *APIClient) (string, error) {
	path, err := api.GetStringValue(
		"/v1/agent/self",
		// format ~ {"config": {"LogFile": "/some/path"}}
		"config", "LogFile",
	)

	if err != nil {
		return "", err
	}
	if path == "" {
		return "", errors.New("empty Nomad LogFile")
	}

	// account for variable behavior depending on destination type
	if _, file := filepath.Split(path); file == "" {
		// this is a directory
		path = path + "nomad*.log"
	} else {
		// this is a "prefix"
		path = strings.Replace(path, ".log", "*.log", 1)
	}

	return path, nil
}
