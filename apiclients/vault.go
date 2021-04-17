package apiclients

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
