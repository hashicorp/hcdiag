package client

// https://www.consul.io/api-docs

import (
	"errors"
	"net/http"
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
func NewConsulAPI() (*APIClient, error) {
	addr := os.Getenv("CONSUL_HTTP_ADDR")
	if addr == "" {
		addr = DefaultConsulAddr
	}

	headers := map[string]string{}
	token := os.Getenv("CONSUL_TOKEN")
	if token != "" {
		headers["X-Consul-Token"] = token
	}

	tlsConfig, err := NewConsulTLSConfig()
	if err != nil {
		return nil, err
	}

	apiClient := NewAPIClient("consul", addr, headers)

	err = configureHttpClientTLS(apiClient.http.(*http.Client), tlsConfig)
	if err != nil {
		return nil, err
	}

	return apiClient, nil
}

// NewConsulTLSConfig returns a *TLSConfig object, using
// default environment variables to build up the object.
func NewConsulTLSConfig() (*TLSConfig, error) {
	tlsConfig := TLSConfig{
		CACert:        os.Getenv(EnvConsulCaCert),
		CAPath:        os.Getenv(EnvConsulCaPath),
		ClientCert:    os.Getenv(EnvConsulClientCert),
		ClientKey:     os.Getenv(EnvConsulClientKey),
		TLSServerName: os.Getenv(EnvConsulTlsServerName),
	}

	if v := os.Getenv(EnvConsulHttpSslVerify); v != "" {
		shouldVerify, err := strconv.ParseBool(v)
		if err != nil {
			return nil, err
		}

		// The semantics Consul uses to indicate whether verification should be done
		// is the opposite of the `tlsConfig.Insecure` field; we negate the value indicated by
		// EnvConsulHttpSslVerify environment variable here to align them.
		tlsConfig.Insecure = !shouldVerify
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
