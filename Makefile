SHELL := /bin/sh
ROOT := $(CURDIR)
export GOCACHE := $(ROOT)/.cache/go-build
export GOMODCACHE := $(ROOT)/.cache/go-mod
export GOTOOLCHAIN := local
export BUF_CACHE_DIR := $(ROOT)/.cache/buf
export PATH := $(ROOT)/.tools/bin:$(PATH)

.PHONY: help deps tidy tools generate fmt build test check test-db basics concurrency up down db migrate migration-status run-catalog run-reading run-gateway
help:
	@echo 'deps tidy tools generate fmt build test check test-db basics concurrency up down db migrate migration-status run-catalog run-reading run-gateway'
deps:
	go mod download
tidy:
	go mod tidy
tools:
	GOBIN=$(ROOT)/.tools/bin go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.12
	GOBIN=$(ROOT)/.tools/bin go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.2
	GOBIN=$(ROOT)/.tools/bin go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.33.0
	GOBIN=$(ROOT)/.tools/bin go install github.com/bufbuild/buf/cmd/buf@v1.50.0
generate:
	$(ROOT)/.tools/bin/buf lint
	$(ROOT)/.tools/bin/buf generate
fmt:
	gofmt -w cmd internal examples migrations
build:
	mkdir -p bin
	go build -o bin/ ./cmd/...
test:
	go test ./...
check:
	go vet ./...
	go test -race ./...
test-db:
	MYSQL_TEST_DSN='school:school@tcp(127.0.0.1:3307)/school?parseTime=true&loc=UTC' go test -race -count=1 -v ./internal/data ./internal/app
basics:
	go run ./examples/01-basics
concurrency:
	go run ./examples/02-concurrency
up:
	docker compose up --build -d --wait
down:
	docker compose down
db:
	docker compose up -d --wait mysql
migrate:
	go run ./cmd/migrate up
migration-status:
	go run ./cmd/migrate status
run-catalog:
	go run ./cmd/catalog
run-reading:
	go run ./cmd/reading
run-gateway:
	go run ./cmd/gateway
