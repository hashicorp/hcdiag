host {
  // Match post-gen
  selects = ["echo dynamic seeker world 1"]
}

product "vault" {
  selects = ["vault debug -output=/VaultDebug.tar.gz -duration=10s"]
}
