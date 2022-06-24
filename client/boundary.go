package client

// https://www.Boundary.io/api-docs

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
)

const (
	DefaultBoundaryAddr = "http://127.0.0.1:8500"

	EnvBoundaryCaCert        = "Boundary_CACERT"
	EnvBoundaryCaPath        = "Boundary_CAPATH"
	EnvBoundaryClientCert    = "Boundary_CLIENT_CERT"
	EnvBoundaryClientKey     = "Boundary_CLIENT_KEY"
	EnvBoundaryHttpSslVerify = "Boundary_HTTP_SSL_VERIFY"
	EnvBoundaryTlsServerName = "Boundary_TLS_SERVER_NAME"
)

// NewBoundaryAPI returns an APIClient for Boundary.
func NewBoundaryAPI() (*APIClient, error) {
	product := "Boundary"

	addr := os.Getenv("Boundary_HTTP_ADDR")
	if addr == "" {
		addr = DefaultBoundaryAddr
	}

	headers := map[string]string{}
	token := os.Getenv("Boundary_TOKEN")
	if token != "" {
		headers["X-Boundary-Token"] = token
	}

	tlsConfig, err := NewBoundaryTLSConfig()
	if err != nil {
		return nil, err
	}

	cfg := APIConfig{
		Product:   product,
		BaseURL:   addr,
		TLSConfig: tlsConfig,
		headers:   headers,
	}

	apiClient, err := NewAPIClient(cfg)
	if err != nil {
		return nil, err
	}

	return apiClient, nil
}

// NewBoundaryTLSConfig returns a TLSConfig object, using
// default environment variables to build up the object.
func NewBoundaryTLSConfig() (TLSConfig, error) {
	// The semantics Boundary uses to indicate whether verification should be done
	// is the opposite of the `tlsConfig.Insecure` field. So, we default shouldVerify
	// to true, determine whether we should actually change it based on the env var,
	// and assign the negation to the `Insecure` field in the return.
	shouldVerify := true

	if v := os.Getenv(EnvBoundaryHttpSslVerify); v != "" {
		var err error
		shouldVerify, err = strconv.ParseBool(v)
		if err != nil {
			return TLSConfig{}, err
		}
	}

	return TLSConfig{
		CACert:        os.Getenv(EnvBoundaryCaCert),
		CAPath:        os.Getenv(EnvBoundaryCaPath),
		ClientCert:    os.Getenv(EnvBoundaryClientCert),
		ClientKey:     os.Getenv(EnvBoundaryClientKey),
		TLSServerName: os.Getenv(EnvBoundaryTlsServerName),
		Insecure:      !shouldVerify,
	}, nil
}

func GetBoundaryLogPath(api *APIClient) (string, error) {
	path, err := api.GetStringValue(
		"/v1/agent/self",
		// format ~ {"DebugConfig": {"Logging": {"LogFilePath": "/some/path"}}}
		"DebugConfig", "Logging", "LogFilePath",
	)

	if err != nil {
		return "", err
	}
	if path == "" {
		return "", errors.New("empty Boundary LogFilePath")
	}

	// account for variable behavior depending on destination type
	if _, file := filepath.Split(path); file == "" {
		// this is a directory
		path = path + "Boundary-*"
	} else {
		// this is a "prefix"
		path = path + "-*"
	}

	return path, nil
}
