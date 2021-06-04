host {
  command {
    run = "testing"
    format = "string"
  }

  GET {
    url = "/v1/api/lol"
  }

  copy {
    path = "./*"
    since = "10h"
  }
}
