# Copyright (c) HashiCorp, Inc.
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
