package client

// https://www.vaultproject.io/api-docs

import (
	"errors"
	"io/ioutil"
	"os"

	home "github.com/mitchellh/go-homedir"
)

const DefaultVaultAddr = "http://127.0.0.1:8200"

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
