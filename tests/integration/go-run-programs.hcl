# used in functional_test.go

program "consul" {
  command = "consul agent -dev"
  check   = "consul members"
}

program "nomad" {
  command = "nomad agent -dev"
  check   = "nomad node status"
  seconds = 60 # windows is extra slow.
}

program "vault" {
  command = "vault server -dev"
  check   = "vault status"
  env = {
    # default client is https, but vault server -dev is http
    VAULT_ADDR = "http://127.0.0.1:8200"
  }
}
