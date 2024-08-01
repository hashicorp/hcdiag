# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: BUSL-1.1

schema = 1
artifacts {
  zip = [
    "hcdiag_${version}_darwin_amd64.zip",
    "hcdiag_${version}_darwin_arm64.zip",
    "hcdiag_${version}_linux_386.zip",
    "hcdiag_${version}_linux_amd64.zip",
    "hcdiag_${version}_linux_arm.zip",
    "hcdiag_${version}_linux_arm64.zip",
    "hcdiag_${version}_windows_386.zip",
    "hcdiag_${version}_windows_amd64.zip",
  ]
  rpm = [
    "hcdiag-${version_linux}-1.aarch64.rpm",
    "hcdiag-${version_linux}-1.armv7hl.rpm",
    "hcdiag-${version_linux}-1.i386.rpm",
    "hcdiag-${version_linux}-1.x86_64.rpm",
  ]
  deb = [
    "hcdiag_${version_linux}-1_amd64.deb",
    "hcdiag_${version_linux}-1_arm64.deb",
    "hcdiag_${version_linux}-1_armhf.deb",
    "hcdiag_${version_linux}-1_i386.deb",
  ]
}
