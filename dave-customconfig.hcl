// bin/hcdiag -autodetect=false -config=dave-customconfig.hcl
agent {
  redact "literal" {
    match = "shazam"
    replace = "custom config (HCL): agent -> redact"
  }
}

host {
  command {
    run = "ps aux | grep hcdiag"
    format = "string"
    redact "literal" {
      match = "fooblarbalurg"
      replace = "custom-config (HCL): host -> command -> redact"
    }
  }
}

product "consul" {
  redact "literal" {
    match = "schmergadillio"
    replace = "custom-config (HCL): product-consul -> redact"
  }

  command {
    run = "consul version"
    format = "json"
  }

  command {
    run = "consul operator raft list-peers"
    format = "json"
  }

  GET {
    path = "/v1/api/metrics?format=prometheus"
  }

  copy {
    path = "/another/test/log"
    since = "240h"
  }

  excludes = ["consul some-verbose-command"]
  selects = ["consul include this", "consul and this"]
}