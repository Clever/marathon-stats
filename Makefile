include golang.mk
.DEFAULT_GOAL := test # override default goal set in library makefile

.PHONY: test build clean vendor $(PKGS)
SHELL := /bin/bash
PKGS := $(shell go list ./... | grep -v /vendor)
$(eval $(call golang-version-check,1.6))

test: $(PKGS)
$(PKGS): golang-test-all-deps
	$(call golang-test-all,$@)

vendor: golang-godep-vendor-deps
	$(call golang-godep-vendor,$(PKGS))

build:
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -ldflags "-X main.Version=$(cat VERSION)" -a -installsuffix cgo

run: build
	HOST=localhost PORT=8082 ./marathon-stats
