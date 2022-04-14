package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	setEnv(t, EnvConsulCaCert, "/tmp/testcerts/ca.crt")
	defer unsetEnv(t, EnvConsulCaCert)

	setEnv(t, EnvConsulCaPath, "/tmp/testcerts/")
	defer unsetEnv(t, EnvConsulCaPath)

	setEnv(t, EnvConsulClientCert, "/tmp/clientcerts/client.crt")
	defer unsetEnv(t, EnvConsulClientCert)

	setEnv(t, EnvConsulClientKey, "/tmp/clientcerts/client.key")
	defer unsetEnv(t, EnvConsulClientKey)

	setEnv(t, EnvConsulTlsServerName, "servername.domain")
	defer unsetEnv(t, EnvConsulTlsServerName)

	setEnv(t, EnvConsulHttpSslVerify, "True")
	defer unsetEnv(t, EnvConsulHttpSslVerify)

	expected := &TLSConfig{
		CACert:        "/tmp/testcerts/ca.crt",
		CAPath:        "/tmp/testcerts/",
		ClientCert:    "/tmp/clientcerts/client.crt",
		ClientKey:     "/tmp/clientcerts/client.key",
		TLSServerName: "servername.domain",
		Insecure:      false,
	}

	actual, err := NewConsulTLSConfig()
	if err != nil {
		t.Errorf("encountered error in NewConsulTLSConfig: %+v", err)
	}

	assert.Equal(t, expected, actual)
}
