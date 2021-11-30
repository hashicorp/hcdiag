run: clean hcdiag
	./hcdiag

run-config: clean hcdiag
	./hcdiag -config test-config.hcl
	$(MAKE) read-results

run-vault-config: clean hcdiag
	./hcdiag -vault -config test-config.hcl
	$(MAKE) read-results

read-results:
	tar xzf hcdiag-*.tar.gz
	cat hcdiag-*/Results.json

hcdiag:
	go build .

test:
	go test ./...

clean:
	rm -rf hcdiag*
