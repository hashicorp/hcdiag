# Redactions

Beginning with version `0.4.0`, `hcdiag` supports redactions. Redactions enable users to tell `hcdiag` about patterns of
text that should be omitted from the results bundle. Redaction matching patterns are based around Regular Expressions.
There are many implementations of Regular Expressions, but `hcdiag` uses the capabilities provided by Golang. Go's
RegEx syntax is the same as the one documented at [https://github.com/google/re2/wiki/Syntax](https://github.com/google/re2/wiki/Syntax),
with the exception that `\C` is not supported.

## Default Redactions

A set of redactions are enabled by default. In version `0.4.0`, these include:

* Email addresses - replaced with `REDACTED@REDACTED` in all products.
* Terraform Enterprise
  * Postgres connection strings are replaced with `postgres://REDACTED`.
  * 32-character alphanumeric strings are replaced with `"REDACTED"`.
  * Key/Value pairs with keys `SECRET=` or `SECRET0=` will have their values redacted, such that
    the result will be `SECRET=REDACTED` or `SECRET0=REDACTED`, respectively.

## Custom Redactions

Users may also specify custom redactions within configuration files, as shown below.
Comments are included for further explanation of how these redactions work. As with any `hcdiag` configuration, be sure
to execute `hcdiag` with the `-config=</the/path/to/your/config-file.hcl>` flag pointed to the proper path so that
redactions will be applied.

```hcl
agent {
  # Redactions in the `agent` block will apply to all products & runners.
  redact "regex" {
    match = "MyPassword"
    # This pattern will match the literal string "MyPassword". Please note that this is still a RegEx, however, so take
    # care to properly escape any special characters that could also be used as RegEx patterns.
  }
}

product "consul" {
  # Redactions in the `product` block will apply to all runners within the product.
  redact "regex" {
    match = "127\\.0\\.0\\.1"
    # This pattern will match the literal string "127.0.0.1"; note that RegEx requires escaping the "." with a backslash.
    # However, HCL requires escaping the backslash with another backslash. This is why double `\`'s are required here.
  }
}

product "tfe" {
  command {
    run = "echo postgres://user:password@xyz.local"
    # This is an unrealistic command example, but it intends to help with demonstration of the following redactor usage.
    format = "string"

    redact "regex" {
    # Redactions inside of a runner block will only apply to this specific runner.
      match = "(postgres://)[^ ]+"
      # This matches any string up to the first space that begins with "postgres://"; the usage of parentheses
      # assigns "postgres://" to a numbered group (group #1). We can use this in a custom replacement.

      replace = "$${1}REDACTED"
      # Here, the `${1}` tells hcdiag to print the contents of group #1, and we follow that with the word "REDACTED".
      # The resulting output will be "postgres://REDACTED" instead of the command output,
      # which would have been "postgres://user:password@xyz.local.
      # Note: Because `${` has a special meaning for HCL, we need to escape this with an additional "$".
    }
  }
}
```
