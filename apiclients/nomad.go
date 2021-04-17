package apiclients

// https://www.nomadproject.io/api-docs

import (
	"os"
)

const DefaultNomadAddr = "http://127.0.0.1:4646"

// NewNomadAPI returns an APIClient for Nomad.
func NewNomadAPI() *APIClient {
	addr := os.Getenv("NOMAD_ADDR")
	if addr == "" {
		addr = DefaultNomadAddr
	}

	headers := map[string]string{}
	token := os.Getenv("NOMAD_TOKEN")
	if token != "" {
		headers["X-Nomad-Token"] = token
	}

	return NewAPIClient("nomad", addr, headers)
}
