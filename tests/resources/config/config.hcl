host {
  command {
    run = "ps aux"
    format = "string"
  }
  copy {
    path = "/var/log/syslog"
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
    path = "/some/test/log"
  }

  copy {
    path = "/another/test/log"
    since = "10d"
  }

  excludes = ["consul some-awfully-long-command"]
  selects = ["consul just this", "consul and this"]
}
