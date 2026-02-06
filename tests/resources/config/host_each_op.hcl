# Copyright IBM Corp. 2021, 2025
# SPDX-License-Identifier: MPL-2.0

host {
  command {
    run = "testing"
    format = "string"
  }

  shell {
    run = "testing"
  }

  GET {
    path = "/v1/api/lol"
  }

  copy {
    path = "./*"
    since = "10h"
  }
}
