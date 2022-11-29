# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

product "consul" {
  command {
    run = "consul version"
    format = "string"
  }
}

product "nomad" {
  command {
    run = "nomad version"
    format = "string"
  }
}

product "vault" {
  command {
    run = "vault version"
    format = "string"
  }
}
