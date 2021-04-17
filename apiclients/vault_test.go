package apiclients

import (
	"os"
	"testing"
)

func TestNewVaultAPI(t *testing.T) {
	os.Setenv("VAULT_TOKEN", "testytest")
	api, err := NewVaultAPI()
	if err != nil {
		t.Error(err)
	}

	if api.Product != "vault" {
		t.Errorf("expected Product to be 'vault'; got '%s'", api.Product)
	}
	if api.BaseURL != DefaultVaultAddr {
		t.Errorf("expected BaseURL to be '%s'; got '%s'", DefaultVaultAddr, api.BaseURL)
	}
	// TODO: test non-default addr, and token
}
