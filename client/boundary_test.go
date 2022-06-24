package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewBoundaryAPI(t *testing.T) {
	api, err := NewBoundaryAPI()
	if err != nil {
		t.Errorf("encountered error from NewBoundaryAPI(); error: %s", err)
	}

	if api.Product != "Boundary" {
		t.Errorf("expected Product to be 'Boundary'; got '%s'", api.Product)
	}
	if api.BaseURL != DefaultBoundaryAddr {
		t.Errorf("expected BaseURL to be '%s'; got '%s'", DefaultBoundaryAddr, api.BaseURL)
	}
	// TODO: test non-default addr, and token
}

func TestGetBoundaryLogPathPDir(t *testing.T) {
	mock := &mockHTTP{
		resp: `{"DebugConfig": {"Logging": {"LogFilePath": "/some/dir/"}}}`,
	}
	api, err := NewBoundaryAPI()
	if err != nil {
		t.Errorf("encountered error from NewBoundaryAPI(); error: %s", err)
	}
	api.http = mock

	path, err := GetBoundaryLogPath(api)
	if err != nil {
		t.Errorf("error running BoundaryLogPath: %s", err)
	}

	expect := "/some/dir/Boundary-*"
	if path != expect {
		t.Errorf("expected LogFilePath='%s'; got: '%s'", expect, path)
	}
}

func TestGetBoundaryLogPathPrefix(t *testing.T) {
	mock := &mockHTTP{
		resp: `{"DebugConfig": {"Logging": {"LogFilePath": "/some/prefix"}}}`,
	}
	api, err := NewBoundaryAPI()
	if err != nil {
		t.Errorf("encountered error from NewBoundaryAPI(); error: %s", err)
	}
	api.http = mock

	path, err := GetBoundaryLogPath(api)
	if err != nil {
		t.Errorf("error running BoundaryLogPath: %s", err)
	}

	expect := "/some/prefix-*"
	if path != expect {
		t.Errorf("expected Boundary LogFilePath='%s'; got: '%s'", expect, path)
	}
}

func TestNewBoundaryTLSConfig(t *testing.T) {
	testCases := []struct {
		name          string
		expectErr     bool
		caCert        string
		caPath        string
		clientCert    string
		clientKey     string
		tlsServerName string
		sslVerify     string
		expected      TLSConfig
	}{
		{
			name:          "Test All Values Set",
			caCert:        "/this_is_not_a_real_location/testcerts/ca.crt",
			caPath:        "/this_is_not_a_real_location/testcerts/",
			clientCert:    "/this_is_not_a_real_location/clientcerts/client.crt",
			clientKey:     "/this_is_not_a_real_location/clientcerts/client.key",
			tlsServerName: "servername.domain",
			sslVerify:     "true",
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
			name:      "Test SSL Verify Set To False Is Insecure",
			sslVerify: "false",
			expected: TLSConfig{
				Insecure: true,
			},
		},
		{
			name:      "Test SSL Verify Not Valid Returns Error",
			sslVerify: "12345",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(EnvBoundaryCaCert, tc.caCert)
			t.Setenv(EnvBoundaryCaPath, tc.caPath)
			t.Setenv(EnvBoundaryClientCert, tc.clientCert)
			t.Setenv(EnvBoundaryClientKey, tc.clientKey)
			t.Setenv(EnvBoundaryTlsServerName, tc.tlsServerName)
			t.Setenv(EnvBoundaryHttpSslVerify, tc.sslVerify)

			actual, err := NewBoundaryTLSConfig()
			if tc.expectErr {
				require.Error(t, err, "an error was expected, but was not raised")
			} else {
				require.NoError(t, err, "encountered unexpected error in NewBoundaryTLSConfig")
				require.Equal(t, tc.expected, actual, "actual TLSConfig does not match the expected struct")
			}
		})
	}
}
