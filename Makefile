.PHONY: test build clean vendor $(PKGS)
SHELL := /bin/bash
PKG = github.com/Clever/marathon-stats
PKGS := $(PKG)

GOVERSION := $(shell go version | grep 1.5)
ifeq "$(GOVERSION)" ""
  $(error must be running Go version 1.5)
endif
export GO15VENDOREXPERIMENT=1

test: $(PKGS)

$(GOPATH)/bin/golint:
	@go get github.com/golang/lint/golint

$(GOPATH)/bin/errcheck:
	@go get github.com/kisielk/errcheck

$(PKGS): $(GOPATH)/bin/golint $(GOPATH)/bin/errcheck
	@echo ""
	@echo "FORMATTING $@..."
	@go get -d -t $@
	@gofmt -w=true $(GOPATH)/src/$@/*.go
	@echo ""
	@echo "LINTING $@..."
	@$(GOPATH)/bin/golint $(GOPATH)/src/$@/*.go
	@echo ""
ifeq ($(COVERAGE),1)
	@echo "TESTING COVERAGE $@..."
	@go test -cover -coverprofile=$(GOPATH)/src/$@/c.out $@ -test.v
	@go tool cover -html=$(GOPATH)/src/$@/c.out
else
	@echo "TESTING $@..."
	@go test -v $@
endif
	@$(GOPATH)/bin/errcheck $@

GODEP := $(GOPATH)/bin/godep
$(GODEP):
	go get -u github.com/tools/godep

vendor: $(GODEP)
	$(GODEP) save $(PKGS)
	find vendor/ -path '*/vendor' -type d | xargs -IX rm -r X # remove any nested vendor directories

build:
	@go build -ldflags "-X main.Version=$(cat VERSION)"

run: build
	HOST=localhost PORT=8082 ./marathon-stats
