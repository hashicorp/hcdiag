# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# tls: https://www.nomadproject.io/docs/configuration/tls
tls {
  http = true
  rpc  = true

  ca_file   = "certs/ca.crt"
  cert_file = "certs/signed.crt"
  key_file  = "certs/signed.key"

  verify_https_client = true
}
