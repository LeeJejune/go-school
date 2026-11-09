SHELL := /bin/sh
ROOT := $(CURDIR)
export GOCACHE := $(ROOT)/.cache/go-build
export GOTOOLCHAIN := local
export GOWORK := off

.PHONY: help basics concurrency test check fmt build
help:
	@echo 'basics concurrency test check fmt build'
basics:
	go run ./examples/01-basics
concurrency:
	go run ./examples/02-concurrency
test:
	go test ./...
check:
	go vet ./...
	go test -race ./...
fmt:
	gofmt -w examples internal
build:
	go build ./...
