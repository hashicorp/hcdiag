package client

// https://www.nomadproject.io/api-docs

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	DefaultNomadAddr = "http://127.0.0.1:4646"

	EnvNomadCaCert        = "NOMAD_CACERT"
	EnvNomadCaPath        = "NOMAD_CAPATH"
	EnvNomadClientCert    = "NOMAD_CLIENT_CERT"
	EnvNomadClientKey     = "NOMAD_CLIENT_KEY"
	EnvNomadSkipVerify    = "NOMAD_SKIP_VERIFY"
	EnvNomadTlsServerName = "NOMAD_TLS_SERVER_NAME"
)

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

// NewNomadTLSConfig returns a *TLSConfig object, using
// default environment variables to build up the object.
func NewNomadTLSConfig() (*TLSConfig, error) {
	tlsConfig := TLSConfig{}
	if v := os.Getenv(EnvNomadCaCert); v != "" {
		tlsConfig.CACert = v
	}

	if v := os.Getenv(EnvNomadCaPath); v != "" {
		tlsConfig.CAPath = v
	}

	if v := os.Getenv(EnvNomadClientCert); v != "" {
		tlsConfig.ClientCert = v
	}

	if v := os.Getenv(EnvNomadClientKey); v != "" {
		tlsConfig.ClientKey = v
	}

	if v := os.Getenv(EnvNomadTlsServerName); v != "" {
		tlsConfig.TLSServerName = v
	}

	if v := os.Getenv(EnvNomadSkipVerify); v != "" {
		skipVerify, err := strconv.ParseBool(v)
		if err != nil {
			return nil, err
		}
		if skipVerify {
			tlsConfig.Insecure = true
		}
	}

	return &tlsConfig, nil
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
		path = strings.Replace(path, ".log", "", 1)
		path = path + "*.log"
	}

	return path, nil
}
