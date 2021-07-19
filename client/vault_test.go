package client

import (
	"os"
	"testing"
)

func TestNewVaultAPI(t *testing.T) {
	os.Setenv("VAULT_TOKEN", "testytest")
	defer os.Unsetenv("VAULT_TOKEN")

	api, err := NewVaultAPI()
	if err != nil {
		t.Errorf("NewVaultAPI: %s", err)
		return
	}

	if api.Product != "vault" {
		t.Errorf("expected Product to be 'vault'; got '%s'", api.Product)
	}
	if api.BaseURL != DefaultVaultAddr {
		t.Errorf("expected BaseURL to be '%s'; got '%s'", DefaultVaultAddr, api.BaseURL)
	}
	// TODO: test non-default addr, and token
}

func TestGetVaultLogPathPDir(t *testing.T) {
	os.Setenv("VAULT_TOKEN", "testytest")
	defer os.Unsetenv("VAULT_TOKEN")

	api, err := NewVaultAPI()
	if err != nil {
		t.Errorf("NewVaultAPI: %s", err)
		return
	}

	mock := &mockHTTP{
		resp: `{"file/": {"options": {"file_path": "/some/log.file"}}}`,
	}
	api.http = mock

	path, err := GetVaultAuditLogPath(api)
	if err != nil {
		t.Errorf("error running VaultLogPath: %s", err)
	}

	expect := "/some/log.file*"
	if path != expect {
		t.Errorf("expected Vault audit log file_path='%s'; got: '%s'", expect, path)
	}
}
