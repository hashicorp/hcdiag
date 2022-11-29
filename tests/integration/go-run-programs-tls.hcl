# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# used in functional_test.go

program "consul" {
  command = "consul agent -dev -config-file=consul.hcl"
  check   = "consul members"
  # env vars: https://www.consul.io/commands#environment-variables=
  env = {
    CONSUL_HTTP_ADDR="https://127.0.0.1:8501"
    CONSUL_CACERT="certs/ca.crt"
    CONSUL_CLIENT_CERT="certs/signed.crt"
    CONSUL_CLIENT_KEY="certs/signed.key"
  }
}

program "nomad" {
  command = "nomad agent -dev -config=nomad.hcl"
  check   = "nomad node status"
  seconds = 60 # windows is extra slow.
  # env vars: https://www.nomadproject.io/docs/commands#mtls-environment-variables=
  env = {
    NOMAD_ADDR="https://127.0.0.1:4646"
    NOMAD_CACERT="certs/ca.crt"
    NOMAD_CLIENT_CERT="certs/signed.crt"
    NOMAD_CLIENT_KEY="certs/signed.key"
  }
}

program "vault" {
  command = "vault server -dev -config=vault.hcl"
  check   = "vault status"
  # env vars: https://www.vaultproject.io/docs/commands
  env = {
    VAULT_ADDR="https://127.0.0.1:8199"
    VAULT_CACERT="certs/ca.crt"
    VAULT_CLIENT_CERT="certs/signed.crt"
    VAULT_CLIENT_KEY="certs/signed.key"
  }
}
