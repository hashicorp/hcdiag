host {
  command {
    run = "testing"
    format = "string"
  }

  GET {
    path = "/v1/api/lol"
  }

  copy {
    path = "./*"
    since = "10h"
  }
}
