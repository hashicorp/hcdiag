# host-diagnostics

## WIP proof of concept / experimentation

### input

* `os` - override operating system value, default is 'auto' and will use runtime.GOOS
* `product` - product cmd execs e.g. vault or terraform

### output

* `results.json` - json output of host diag + `results.tar`, `results.tar.gz`
* `results_product.json` - json output of prod cmd exec if product entered

### run, build, and test

host diag  
* `go run .`

host diag & vault cmd  
* `go run . -product vault`

linux build  
* `GOOS=linux GOARCH=amd64 go build -o host-diagnostics_linux`

testing ubuntu and rhel  
* `cd ./tests && ./testing.sh`
