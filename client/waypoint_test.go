package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewWaypointAPI(t *testing.T) {
	api, err := NewWaypointAPI()
	if err != nil {
		t.Errorf("encountered error from NewWaypointAPI(); error: %s", err)
	}

	if api.Product != "waypoint" {
		t.Errorf("expected Product to be 'waypoint'; got '%s'", api.Product)
	}
	if api.BaseURL != DefaultWaypointAddr {
		t.Errorf("expected BaseURL to be '%s'; got '%s'", DefaultWaypointAddr, api.BaseURL)
	}
	// TODO: test non-default addr, and token
}

func TestGetWaypointLogPathPDir(t *testing.T) {
	mock := &mockHTTP{
		resp: `{"DebugConfig": {"Logging": {"LogFilePath": "/some/dir/"}}}`,
	}
	api, err := NewWaypointAPI()
	if err != nil {
		t.Errorf("encountered error from NewWaypointAPI(); error: %s", err)
	}
	api.http = mock

	path, err := GetWaypointLogPath(api)
	if err != nil {
		t.Errorf("error running WaypointLogPath: %s", err)
	}

	expect := "/some/dir/waypoint-*"
	if path != expect {
		t.Errorf("expected LogFilePath='%s'; got: '%s'", expect, path)
	}
}

func TestGetWaypointLogPathPrefix(t *testing.T) {
	mock := &mockHTTP{
		resp: `{"DebugConfig": {"Logging": {"LogFilePath": "/some/prefix"}}}`,
	}
	api, err := NewWaypointAPI()
	if err != nil {
		t.Errorf("encountered error from NewWaypointAPI(); error: %s", err)
	}
	api.http = mock

	path, err := GetWaypointLogPath(api)
	if err != nil {
		t.Errorf("error running WaypointLogPath: %s", err)
	}

	expect := "/some/prefix-*"
	if path != expect {
		t.Errorf("expected Waypoint LogFilePath='%s'; got: '%s'", expect, path)
	}
}

func TestNewWaypointTLSConfig(t *testing.T) {
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
			t.Setenv(EnvWaypointCaCert, tc.caCert)
			t.Setenv(EnvWaypointCaPath, tc.caPath)
			t.Setenv(EnvWaypointClientCert, tc.clientCert)
			t.Setenv(EnvWaypointClientKey, tc.clientKey)
			t.Setenv(EnvWaypointTlsServerName, tc.tlsServerName)
			t.Setenv(EnvWaypointHttpSslVerify, tc.sslVerify)

			actual, err := NewWaypointTLSConfig()
			if tc.expectErr {
				require.Error(t, err, "an error was expected, but was not raised")
			} else {
				require.NoError(t, err, "encountered unexpected error in NewWaypointTLSConfig")
				require.Equal(t, tc.expected, actual, "actual TLSConfig does not match the expected struct")
			}
		})
	}
}
