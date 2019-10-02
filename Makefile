.PHONY: default test build

OS := $(shell uname)
VERSION ?= 1.0.0

PKG_NAME = github.com/AdhityaRamadhanus/chronicle

# target #

default: unit-test build

# Test Packages

unit-test:
	go test -count=1 -v -short --cover ./...

integration-test:
	go test -count=1 -run Integration -v --cover ./...

migration:
	go run script/run_migration/main.go
