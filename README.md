# host-diagnostics

## WIP proof of concept / experimentation

### input

* `os` - string - override operating system value, `auto` by default and will use runtime.GOOS
* `consul` - bool - run consul diagnostics
* `nomad` - bool - run noamd diagnostics
* `vault` - bool - run vault diagnostics
* `all` - bool - run all available product diagnostics
* `dryrun` - bool - if true will display but not execute os and product cmd's
* `outfile` - string - name of output file, `support.tar.gz` by default
* `includeDir` - string - include a directory in output bundle (e.g. logs)
* `includeFile` - string - include a file in output bundle

### output

* `support.tar.gz` - contents include Results.json, Manifest.json, and product diagnostic outputs where applicable

### run, build, and test

host diag only  
* `go run .`

host diag & vault diagnostics  
* `go run . -vault`

host diag & consul + nomad diagnostics  
* `go run . -consul -nomad`
