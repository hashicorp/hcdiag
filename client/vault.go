package client

// https://www.vaultproject.io/api-docs

import (
	"errors"
	"io/ioutil"
	"os"
	"strconv"

	home "github.com/mitchellh/go-homedir"
)

const (
	DefaultVaultAddr = "http://127.0.0.1:8200"

	EnvVaultCaCert        = "VAULT_CACERT"
	EnvVaultCaPath        = "VAULT_CAPATH"
	EnvVaultClientCert    = "VAULT_CLIENT_CERT"
	EnvVaultClientKey     = "VAULT_CLIENT_KEY"
	EnvVaultSkipVerify    = "VAULT_SKIP_VERIFY"
	EnvVaultTlsServerName = "VAULT_TLS_SERVER_NAME"
)

// NewVaultAPI returns an APIClient for Vault.
func NewVaultAPI() (*APIClient, error) {
	addr := os.Getenv("VAULT_ADDR")
	if addr == "" {
		addr = DefaultVaultAddr
	}

	headers := map[string]string{}
	token := os.Getenv("VAULT_TOKEN")
	switch token {
	case "":
		path, err := home.Expand("~/.vault-token")
		if err != nil {
			break
		}
		bts, err := ioutil.ReadFile(path)
		if err != nil {
			break
		}
		token = string(bts)
	}
	if token == "" {
		return nil, errors.New("unable to find VAULT_TOKEN env or ~/.vault-token")
	}
	headers["X-Vault-Token"] = token

	return NewAPIClient("vault", addr, headers), nil
}

// NewVaultTLSConfig returns a *TLSConfig object, using
// default environment variables to build up the object.
func NewVaultTLSConfig() (*TLSConfig, error) {
	tlsConfig := TLSConfig{}
	if v := os.Getenv(EnvVaultCaCert); v != "" {
		tlsConfig.CACert = v
	}

	if v := os.Getenv(EnvVaultCaPath); v != "" {
		tlsConfig.CAPath = v
	}

	if v := os.Getenv(EnvVaultClientCert); v != "" {
		tlsConfig.ClientCert = v
	}

	if v := os.Getenv(EnvVaultClientKey); v != "" {
		tlsConfig.ClientKey = v
	}

	if v := os.Getenv(EnvVaultTlsServerName); v != "" {
		tlsConfig.TLSServerName = v
	}

	if v := os.Getenv(EnvVaultSkipVerify); v != "" {
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

func GetVaultAuditLogPath(api *APIClient) (string, error) {
	// this assumes operators have not enabled a custom audit -path
	// i.e. only the first example here: https://www.vaultproject.io/docs/audit/file
	// $ vault audit enable file file_path=/some/log/file.log
	// will be found.

	path, err := api.GetStringValue(
		"/v1/sys/audit",
		// format ~ {"file/": {"options": {"file_path": "/some/path"}}}
		"file/", "options", "file_path",
	)

	if err != nil {
		return path, err
	}
	if path == "" {
		return path, errors.New("empty Vault audit log file_path")
	}

	path = path + "*" // for if they use `logrotate` or similar

	return path, nil
}
