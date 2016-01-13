.PHONY: test build clean $(PKGS)
SHELL := /bin/bash
PKG = github.com/Clever/marathon-stats
PKGS := $(PKG)

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

build:
	@go build -ldflags "-X main.Version=$(cat VERSION)"
