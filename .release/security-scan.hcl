# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

binary {
  secrets      = true
  go_modules   = true
  osv          = true
  oss_index    = false
  nvd          = false

  triage {
    suppress {
      vulnerabilities = [
        // GO-2026-5932 concerns unmaintained x/crypto/openpgp packages;
        // its OSV affected range has no fixed version. CRT reports it at
        // module level, but hcdiag's build dependency graphs exclude
        // openpgp and all its subpackages.
        // Verified at cc2bf42 with Go 1.26.10 using
        // `go list -mod=readonly -deps -json .`, CGO_ENABLED=0, for every
        // GOOS/GOARCH target in .github/workflows/build.yml (8 targets).
        // Only x/crypto/{bcrypt,blowfish,pbkdf2,scrypt} are included.
        // Suppress this non-applicable finding, not an accepted OpenPGP risk.
        // Revalidate when dependencies or release targets change; remove
        // this suppression if any affected package enters the build graph.
        // https://github.com/hashicorp/hcdiag/security/code-scanning/141
        // Other examples of this vulnerability being suppressed:
        // https://github.com/hashicorp/tfctl-cli/blob/main/.release/security-scan.hcl#L17-L27
        "GO-2026-5932",
      ]
    }
  }
}
