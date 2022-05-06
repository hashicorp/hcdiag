# this is here to conveniently generate a self-signed CA and use it to sign a cert/key pair.

locals {
  valid_hours = 24 * 365 * 50 // 50 years

  files = {
    "ca.crt"     = tls_self_signed_cert.ca.cert_pem
    "signed.key" = tls_private_key.crt.private_key_pem
    "signed.crt" = tls_locally_signed_cert.crt.cert_pem
  }
}

output "files" {
  value = keys(local.files)
}

resource "local_file" "certs" {
  for_each = local.files
  content  = each.value
  filename = "${path.module}/${each.key}"
}

resource "tls_private_key" "ca" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "ca" {
  private_key_pem = tls_private_key.ca.private_key_pem

  is_ca_certificate     = true
  validity_period_hours = local.valid_hours

  subject {
    common_name         = "hcdiag-test"
    organization        = "HashiCorp"
    organizational_unit = "CORI"
  }

  allowed_uses = [
    "cert_signing",
    "crl_signing",
  ]
}

resource "tls_private_key" "crt" {
  algorithm = "RSA"
}

resource "tls_cert_request" "csr" {
  private_key_pem = tls_private_key.crt.private_key_pem

  subject {
    common_name = "hcdiag-test"
  }

  dns_names    = ["localhost"]
  ip_addresses = ["127.0.0.1"]
}

resource "tls_locally_signed_cert" "crt" {
  ca_cert_pem        = tls_self_signed_cert.ca.cert_pem
  ca_private_key_pem = tls_private_key.ca.private_key_pem
  cert_request_pem   = tls_cert_request.csr.cert_request_pem

  validity_period_hours = local.valid_hours

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "client_auth",
    "server_auth",
  ]
}
