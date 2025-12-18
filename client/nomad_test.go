// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewNomadAPI(t *testing.T) {
	api, err := NewNomadAPI()
	if err != nil {
		t.Errorf("encountered error from NewNomadAPI(); error: %s", err)
	}

	if api.Product != "nomad" {
		t.Errorf("expected Product to be 'nomad'; got '%s'", api.Product)
	}
	if api.BaseURL != DefaultNomadAddr {
		t.Errorf("expected BaseURL to be '%s'; got '%s'", DefaultNomadAddr, api.BaseURL)
	}
	// TODO: test non-default addr, and token
}

func TestGetNomadLogPathPDir(t *testing.T) {
	mock := &mockHTTP{
		resp: `{"config": {"LogFile": "/some/dir/"}}`,
	}
	api, err := NewNomadAPI()
	if err != nil {
		t.Errorf("encountered error from NewNomadAPI(); error: %s", err)
	}
	api.http = mock

	path, err := GetNomadLogPath(api)
	if err != nil {
		t.Errorf("error running NomadLogPath: %s", err)
	}

	expect := "/some/dir/nomad*.log"
	if path != expect {
		t.Errorf("expected LogFile='%s'; got: '%s'", expect, path)
	}
}

func TestGetNomadLogPathPrefix(t *testing.T) {
	mock := &mockHTTP{
		resp: `{"config": {"LogFile": "/some/prefix"}}`,
	}
	api, err := NewNomadAPI()
	if err != nil {
		t.Errorf("encountered error from NewNomadAPI(); error: %s", err)
	}
	api.http = mock

	path, err := GetNomadLogPath(api)
	if err != nil {
		t.Errorf("error running NomadLogPath: %s", err)
	}

	expect := "/some/prefix*.log"
	if path != expect {
		t.Errorf("expected Nomad LogFile='%s'; got: '%s'", expect, path)
	}
}

func TestNewNomadTLSConfig(t *testing.T) {
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
			t.Setenv(EnvNomadCaCert, tc.caCert)
			t.Setenv(EnvNomadCaPath, tc.caPath)
			t.Setenv(EnvNomadClientCert, tc.clientCert)
			t.Setenv(EnvNomadClientKey, tc.clientKey)
			t.Setenv(EnvNomadTlsServerName, tc.tlsServerName)
			t.Setenv(EnvNomadSkipVerify, tc.skipVerify)

			actual, err := NewNomadTLSConfig()
			if tc.expectErr {
				require.Error(t, err, "an error was expected, but was not raised")
			} else {
				require.NoError(t, err, "encountered unexpected error in NewConsulTLSConfig")
				require.Equal(t, tc.expected, actual, "actual TLSConfig does not match the expected struct")
			}
		})
	}
}
