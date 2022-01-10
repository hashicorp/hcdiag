host {
  command {
    run = "testing"
    format = "string"
  }

  shell {
    run = "testing"
  }

  GET {
    path = "/v1/api/lol"
  }

  copy {
    path = "./*"
    since = "10h"
  }
}
