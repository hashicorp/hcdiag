package apiclients

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
