package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConsulAPI(t *testing.T) {
	api, err := NewConsulAPI()
	if err != nil {
		t.Errorf("encountered error from NewConsulAPI(); error: %s", err)
	}

	if api.Product != "consul" {
		t.Errorf("expected Product to be 'consul'; got '%s'", api.Product)
	}
	if api.BaseURL != DefaultConsulAddr {
		t.Errorf("expected BaseURL to be '%s'; got '%s'", DefaultConsulAddr, api.BaseURL)
	}
	// TODO: test non-default addr, and token
}

func TestGetConsulLogPathPDir(t *testing.T) {
	mock := &mockHTTP{
		resp: `{"DebugConfig": {"Logging": {"LogFilePath": "/some/dir/"}}}`,
	}
	api, err := NewConsulAPI()
	if err != nil {
		t.Errorf("encountered error from NewConsulAPI(); error: %s", err)
	}
	api.http = mock

	path, err := GetConsulLogPath(api)
	if err != nil {
		t.Errorf("error running ConsulLogPath: %s", err)
	}

	expect := "/some/dir/consul-*"
	if path != expect {
		t.Errorf("expected LogFilePath='%s'; got: '%s'", expect, path)
	}
}

func TestGetConsulLogPathPrefix(t *testing.T) {
	mock := &mockHTTP{
		resp: `{"DebugConfig": {"Logging": {"LogFilePath": "/some/prefix"}}}`,
	}
	api, err := NewConsulAPI()
	if err != nil {
		t.Errorf("encountered error from NewConsulAPI(); error: %s", err)
	}
	api.http = mock

	path, err := GetConsulLogPath(api)
	if err != nil {
		t.Errorf("error running ConsulLogPath: %s", err)
	}

	expect := "/some/prefix-*"
	if path != expect {
		t.Errorf("expected Consul LogFilePath='%s'; got: '%s'", expect, path)
	}
}

func TestNewConsulTLSConfig(t *testing.T) {
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
			t.Setenv(EnvConsulCaCert, tc.caCert)
			t.Setenv(EnvConsulCaPath, tc.caPath)
			t.Setenv(EnvConsulClientCert, tc.clientCert)
			t.Setenv(EnvConsulClientKey, tc.clientKey)
			t.Setenv(EnvConsulTlsServerName, tc.tlsServerName)
			t.Setenv(EnvConsulHttpSslVerify, tc.sslVerify)

			actual, err := NewConsulTLSConfig()
			if tc.expectErr {
				require.Error(t, err, "an error was expected, but was not raised")
			} else {
				require.NoError(t, err, "encountered unexpected error in NewConsulTLSConfig")
				require.NotNil(t, actual, "expected output object to not be nil")
				require.Equal(t, tc.expected, *actual, "actual TLSConfig does not match the expected struct")
			}
		})
	}
}
