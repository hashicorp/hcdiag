# Copyright IBM Corp. 2021, 2025
# SPDX-License-Identifier: MPL-2.0

host {
  command {
    run = "testing"
    format = "string"
  }

  command {
    run = "another one"
    format = "string"
  }

  command {
    run = "do a thing"
    format = "json"
  }
}
