# Copyright IBM Corp. 2021, 2025
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
