# host-diagnostics

## WIP proof of concept / experimentation

### input

* `os` - string - override operating system value, `auto` by default and will use runtime.GOOS
* `product` - string - product cmd execs e.g. vault or terraform, blank string by default
* `dryrun` - bool - if true will display but not execute os and product cmd's, `false` by default
* `outfile` - string - name of output file, `support.tar.gz` by default

### output

* `support.tar.gz` - contents include HostInfo.json and ProductInfo.json

### run, build, and test

host diag  
* `go run .`

host diag & vault cmd  
* `go run . -product vault`

linux build  
* `GOOS=linux GOARCH=amd64 go build -o host-diagnostics_linux`

testing ubuntu and rhel  
* `cd ./tests && ./testing.sh`
