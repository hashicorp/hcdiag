storage "inmem" {}

# listener: https://www.vaultproject.io/docs/configuration/listener/tcp
listener "tcp" {
  address       = "0.0.0.0:8199" # 8200 is http with `vault server -dev`

  tls_client_ca_file = "certs/ca.crt"
  tls_cert_file      = "certs/signed.crt"
  tls_key_file       = "certs/signed.key"

  tls_require_and_verify_client_cert = true
}
