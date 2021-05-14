package apiclients

// https://www.consul.io/api-docs

import (
	"errors"
	"os"
	"path/filepath"
)

const DefaultConsulAddr = "http://127.0.0.1:8500"

// NewConsulAPI returns an APIClient for Consul.
func NewConsulAPI() *APIClient {
	addr := os.Getenv("CONSUL_HTTP_ADDR")
	if addr == "" {
		addr = DefaultConsulAddr
	}

	headers := map[string]string{}
	token := os.Getenv("CONSUL_TOKEN")
	if token != "" {
		headers["X-Consul-Token"] = token
	}

	return NewAPIClient("consul", addr, headers)
}

func GetConsulLogPath(api *APIClient) (string, error) {
	path, err := api.GetStringValue(
		"/v1/agent/self",
		// format ~ {"DebugConfig": {"Logging": {"LogFilePath": "/some/path"}}}
		"DebugConfig", "Logging", "LogFilePath",
	)

	if err != nil {
		return "", err
	}
	if path == "" {
		return "", errors.New("empty Consul LogFilePath")
	}

	// account for variable behavior depending on destination type
	if _, file := filepath.Split(path); file == "" {
		// this is a directory
		path = path + "consul-*"
	} else {
		// this is a "prefix"
		path = path + "-*"
	}

	return path, nil
}
