// bin/hcdiag -autodetect=false -config=dave-customconfig.hcl
agent {
  redact "literal" {
    match = "agent"
    replace = "l'agente, c'est bon!"
  }
}

host {
  command {
    run = "ps aux | grep hcdiag"
    format = "string"
    redact "literal" {
      match = "hcdiag"
      replace = "fooblarbalurg"
    }
  }
}

product "consul" {
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