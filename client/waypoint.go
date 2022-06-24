package client

// https://www.waypoint.io/api-docs

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
)

const (
	DefaultWaypointAddr = "http://127.0.0.1:8500"

	EnvWaypointCaCert        = "CONSUL_CACERT"
	EnvWaypointCaPath        = "CONSUL_CAPATH"
	EnvWaypointClientCert    = "CONSUL_CLIENT_CERT"
	EnvWaypointClientKey     = "CONSUL_CLIENT_KEY"
	EnvWaypointHttpSslVerify = "CONSUL_HTTP_SSL_VERIFY"
	EnvWaypointTlsServerName = "CONSUL_TLS_SERVER_NAME"
)

// NewWaypointAPI returns an APIClient for Waypoint.
func NewWaypointAPI() (*APIClient, error) {
	product := "waypoint"

	addr := os.Getenv("CONSUL_HTTP_ADDR")
	if addr == "" {
		addr = DefaultWaypointAddr
	}

	headers := map[string]string{}
	token := os.Getenv("CONSUL_TOKEN")
	if token != "" {
		headers["X-Waypoint-Token"] = token
	}

	tlsConfig, err := NewWaypointTLSConfig()
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

// NewWaypointTLSConfig returns a TLSConfig object, using
// default environment variables to build up the object.
func NewWaypointTLSConfig() (TLSConfig, error) {
	// The semantics Waypoint uses to indicate whether verification should be done
	// is the opposite of the `tlsConfig.Insecure` field. So, we default shouldVerify
	// to true, determine whether we should actually change it based on the env var,
	// and assign the negation to the `Insecure` field in the return.
	shouldVerify := true

	if v := os.Getenv(EnvWaypointHttpSslVerify); v != "" {
		var err error
		shouldVerify, err = strconv.ParseBool(v)
		if err != nil {
			return TLSConfig{}, err
		}
	}

	return TLSConfig{
		CACert:        os.Getenv(EnvWaypointCaCert),
		CAPath:        os.Getenv(EnvWaypointCaPath),
		ClientCert:    os.Getenv(EnvWaypointClientCert),
		ClientKey:     os.Getenv(EnvWaypointClientKey),
		TLSServerName: os.Getenv(EnvWaypointTlsServerName),
		Insecure:      !shouldVerify,
	}, nil
}

func GetWaypointLogPath(api *APIClient) (string, error) {
	path, err := api.GetStringValue(
		"/v1/agent/self",
		// format ~ {"DebugConfig": {"Logging": {"LogFilePath": "/some/path"}}}
		"DebugConfig", "Logging", "LogFilePath",
	)

	if err != nil {
		return "", err
	}
	if path == "" {
		return "", errors.New("empty Waypoint LogFilePath")
	}

	// account for variable behavior depending on destination type
	if _, file := filepath.Split(path); file == "" {
		// this is a directory
		path = path + "waypoint-*"
	} else {
		// this is a "prefix"
		path = path + "-*"
	}

	return path, nil
}
