package client

import (
	"testing"
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
