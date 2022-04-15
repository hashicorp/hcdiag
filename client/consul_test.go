package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConsulAPI(t *testing.T) {
	api := NewConsulAPI()

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
	api := NewConsulAPI()
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
	api := NewConsulAPI()
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
			name:          "TestAllValuesSet",
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
			name:     "TestNoValuesSet",
			expected: TLSConfig{},
		},
		{
			name:      "TestSslVerifySetToFalseIsInsecure",
			sslVerify: "false",
			expected: TLSConfig{
				Insecure: true,
			},
		},
		{
			name:      "TestSslVerifyNotBoolean",
			sslVerify: "12345",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.caCert != "" {
				t.Setenv(EnvConsulCaCert, tc.caCert)
			}

			if tc.caPath != "" {
				t.Setenv(EnvConsulCaPath, tc.caPath)
			}

			if tc.clientCert != "" {
				t.Setenv(EnvConsulClientCert, tc.clientCert)
			}

			if tc.clientKey != "" {
				t.Setenv(EnvConsulClientKey, tc.clientKey)
			}

			if tc.tlsServerName != "" {
				t.Setenv(EnvConsulTlsServerName, tc.tlsServerName)
			}

			if tc.sslVerify != "" {
				t.Setenv(EnvConsulHttpSslVerify, tc.sslVerify)
			}

			actual, err := NewConsulTLSConfig()
			if tc.expectErr {
				require.Error(t, err, "an error was expected, but was not raised")
			} else {
				require.NoError(t, err, "encountered unexpected error in NewConsulTLSConfig")
				require.Equal(t, tc.expected, *actual, "actual TLSConfig does not match the expected struct")
			}
		})
	}
}
