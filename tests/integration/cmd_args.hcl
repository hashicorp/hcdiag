host {
  command {
    run    = "jq -n '$in.\"foo bar\"' --argjson in '{\"foo bar\": 22}'"
    format = "string"
  }
}
