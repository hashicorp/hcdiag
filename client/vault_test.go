// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
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

func TestNewVaultTLSConfig(t *testing.T) {
	testCases := []struct {
		name          string
		expectErr     bool
		caCert        string
		caPath        string
		clientCert    string
		clientKey     string
		tlsServerName string
		skipVerify    string
		expected      TLSConfig
	}{
		{
			name:          "Test All Values Set",
			caCert:        "/this_is_not_a_real_location/testcerts/ca.crt",
			caPath:        "/this_is_not_a_real_location/testcerts/",
			clientCert:    "/this_is_not_a_real_location/clientcerts/client.crt",
			clientKey:     "/this_is_not_a_real_location/clientcerts/client.key",
			tlsServerName: "servername.domain",
			skipVerify:    "false",
			expected: TLSConfig{
				CACert:        "/this_is_not_a_real_location/testcerts/ca.crt",
				CAPath:        "/this_is_not_a_real_location/testcerts/",
				ClientCert:    "/this_is_not_a_real_location/clientcerts/client.crt",
				ClientKey:     "/this_is_not_a_real_location/clientcerts/client.key",
				TLSServerName: "servername.domain",
				Insecure:      false,
			},
		},
		{
			name:     "Test No Values Set",
			expected: TLSConfig{},
		},
		{
			name:       "Test Skip Verify Set To False Is Insecure",
			skipVerify: "true",
			expected: TLSConfig{
				Insecure: true,
			},
		},
		{
			name:       "Test Skip Verify Not Boolean Returns Error",
			skipVerify: "12345",
			expectErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(EnvVaultCaCert, tc.caCert)
			t.Setenv(EnvVaultCaPath, tc.caPath)
			t.Setenv(EnvVaultClientCert, tc.clientCert)
			t.Setenv(EnvVaultClientKey, tc.clientKey)
			t.Setenv(EnvVaultTlsServerName, tc.tlsServerName)
			t.Setenv(EnvVaultSkipVerify, tc.skipVerify)

			actual, err := NewVaultTLSConfig()
			if tc.expectErr {
				require.Error(t, err, "an error was expected, but was not raised")
			} else {
				require.NoError(t, err, "encountered unexpected error in NewConsulTLSConfig")
				require.Equal(t, tc.expected, actual, "actual TLSConfig does not match the expected struct")
			}
		})
	}
}
