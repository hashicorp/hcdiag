# tls: https://www.consul.io/docs/agent/options#tls-configuration-reference=

addresses = {
  https = "0.0.0.0"
}

ports {
  https = 8501 # recommended to use 8501 if enabling tls
  grpc_tls = 8503 # TLS port must be set beginning in Consul 1.14.0
  http  = 8888 # change from default to expose any breakage from clients' default addrs
}

ca_file   = "certs/ca.crt"
cert_file = "certs/signed.crt"
key_file  = "certs/signed.key"

verify_incoming_https = true
