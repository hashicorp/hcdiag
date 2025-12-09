# Copyright IBM Corp. 2021, 2025
# SPDX-License-Identifier: MPL-2.0

host {
  command {
    run = "ps aux"
    format = "string"
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

  shell {
    run = "consul members | grep ."
  }

  GET {
    path = "/v1/api/metrics?format=prometheus"
  }

  copy {
    path = "/another/test/log"
    since = "240h"
  }

  excludes = ["consul some-awfully-long-command"]
  selects = ["consul just this", "consul and this"]
}
