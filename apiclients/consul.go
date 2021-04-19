package apiclients

// https://www.consul.io/api-docs

import (
	"os"
)

const DefaultConsulAddr = "http://127.0.0.1:8500"

// NewConsulAPI returns an APIClient for Consul.
func NewConsulAPI() *APIClient {
	addr := os.Getenv("CONSUL_HTTP_ADDR")
	if addr == "" {
		addr = DefaultConsulAddr
	}

	headers := map[string]string{}
	token := os.Getenv("CONSUL_TOKEN")
	if token != "" {
		headers["X-Consul-Token"] = token
	}

	return NewAPIClient("consul", addr, headers)
}
