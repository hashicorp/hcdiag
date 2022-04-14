package client

// https://www.consul.io/api-docs

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
)

const (
	DefaultConsulAddr = "http://127.0.0.1:8500"

	EnvConsulCaCert        = "CONSUL_CACERT"
	EnvConsulCaPath        = "CONSUL_CAPATH"
	EnvConsulClientCert    = "CONSUL_CLIENT_CERT"
	EnvConsulClientKey     = "CONSUL_CLIENT_KEY"
	EnvConsulHttpSslVerify = "CONSUL_HTTP_SSL_VERIFY"
	EnvConsulTlsServerName = "CONSUL_TLS_SERVER_NAME"
)

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

func NewConsulTLSConfig() (*TLSConfig, error) {
	tlsConfig := TLSConfig{}
	if v := os.Getenv(EnvConsulCaCert); v != "" {
		tlsConfig.CACert = v
	}

	if v := os.Getenv(EnvConsulCaPath); v != "" {
		tlsConfig.CAPath = v
	}

	if v := os.Getenv(EnvConsulClientCert); v != "" {
		tlsConfig.ClientCert = v
	}

	if v := os.Getenv(EnvConsulClientKey); v != "" {
		tlsConfig.ClientKey = v
	}

	if v := os.Getenv(EnvConsulTlsServerName); v != "" {
		tlsConfig.TLSServerName = v
	}

	if v := os.Getenv(EnvConsulHttpSslVerify); v != "" {
		doVerify, err := strconv.ParseBool(v)
		if err != nil {
			return nil, err
		}
		if !doVerify {
			tlsConfig.Insecure = true
		}
	}

	return &tlsConfig, nil
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
