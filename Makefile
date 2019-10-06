.PHONY: default test build

OS := $(shell uname)
VERSION ?= 1.0.0

# target #

default: unit-test build

# Test Packages

unit-test:
	go test -count=1 -v --cover ./... -tags="unit"

integration-test:
	@go test ./... -count=1 -v --cover -tags="authentication" | { grep -v 'no test files'; true; }
	@go test ./... -count=1 -v --cover -tags="profile" | { grep -v 'no test files'; true; }
	@go test ./... -count=1 -v --cover -tags="repository" | { grep -v 'no test files'; true; }

migration:
	go run script/run_migration/main.go
