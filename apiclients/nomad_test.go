package apiclients

import (
	"testing"
)

func TestNewNomadAPI(t *testing.T) {
	api := NewNomadAPI()

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
	api := NewNomadAPI()
	api.http = mock

	path, err := GetNomadLogPath(api)
	if err != nil {
		t.Errorf("error running NomadLogPath: %s", err)
	}

	expect := "/some/dir/nomad-*"
	if path != expect {
		t.Errorf("expected LogFile='%s'; got: '%s'", expect, path)
	}
}

func TestGetNomadLogPathPrefix(t *testing.T) {
	mock := &mockHTTP{
		resp: `{"config": {"LogFile": "/some/prefix"}}`,
	}
	api := NewNomadAPI()
	api.http = mock

	path, err := GetNomadLogPath(api)
	if err != nil {
		t.Errorf("error running NomadLogPath: %s", err)
	}

	expect := "/some/prefix-*"
	if path != expect {
		t.Errorf("expected Nomad LogFile='%s'; got: '%s'", expect, path)
	}
}
