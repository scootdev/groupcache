NAME := groupcache
DESC := caching library
GO111MODULE := on
export GO111MODULE

SHELL := /bin/bash -o pipefail

default:
	go build ./...

format:
	go fmt ./...

vet:
	go vet ./...

test:
	go test -race ./...

coverage:
	sh testCoverage.sh

clean:
	go clean ./...

ci: format vet test coverage
