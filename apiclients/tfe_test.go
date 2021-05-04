package apiclients

import (
	"testing"
)

func TestNewTFEAPI(t *testing.T) {
	api := NewTFEAPI()

	if api.Product != "tfe" {
		t.Errorf("expected Product to be 'tfe'; got '%s'", api.Product)
	}
	if api.BaseURL != DefaultTFEAddr {
		t.Errorf("expected BaseURL to be '%s'; got '%s'", DefaultTFEAddr, api.BaseURL)
	}
	// TODO: test non-default addr, and token
}
