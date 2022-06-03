# don't cache tests
export GOFLAGS = -count=1

VERSION := $(shell ./build-scripts/version.sh version/version.go)

GIT_COMMIT := $(shell git rev-parse --short HEAD)
GIT_DIRTY := $(if $(shell git status --porcelain),+CHANGES)
BUILD_DATE := $(shell date)

COMMIT_FLAG := -X 'github.com/hashicorp/hcdiag/version.gitCommit=$(GIT_COMMIT)$(GIT_DIRTY)'
BUILD_DATE_FLAG := -X 'github.com/hashicorp/hcdiag/version.buildDate=$(BUILD_DATE)'
GO_LDFLAGS := "-s -w ${COMMIT_FLAG} ${BUILD_DATE_FLAG}"

help: ## show this make help
	@awk -F'[:#]' '/#\#/ { printf "%-15s %s\n", $$1, $$NF }' $(MAKEFILE_LIST)
.PHONY: help

env: ## env vars; eval $(make env)
	@echo "$(PATH)" | grep -q "$(PWD)/bin" || echo 'export PATH=$$PWD/bin:$$PATH'
	@echo 'export VAULT_SKIP_VERIFY=1'
.PHONY: env

build: bin/hcdiag ## build bin/hcdiag

bin:
	mkdir -p bin

bin/hcdiag: bin
	go build -trimpath -ldflags=$(GO_LDFLAGS) -o bin .

test: ## run tests
	go test -cover ./...
.PHONY: test

test-functional: bin/hcdiag show-versions ## run functional tests
	go test -v ./tests/integration/ -tags=functional
.PHONY: test-functional

show-versions: ## show product and hcdiag versions
	which consul && consul version
	@echo
	which nomad && nomad version
	@echo
	which vault && vault version
	@echo
	which hcdiag && hcdiag -version
	@echo
.PHONY: show-versions

clean: ## clean bin and bundle files
	rm -rf bin/ hcdiag-*
.PHONY: clean

version: ## Show the version of the project
	@echo $(VERSION)
.PHONY: version

# windows:
# $env:path = "$pwd/bin;$env:path"
