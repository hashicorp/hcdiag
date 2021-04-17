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
