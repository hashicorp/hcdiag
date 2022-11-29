# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

product "consul" {
  excludes = ["*debug*"]
}
product "nomad" {
  excludes = ["*debug*"]
}
product "vault" {
  excludes = ["*debug*"]
}
