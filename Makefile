.PHONY: default test build

OS := $(shell uname)
VERSION ?= 1.0.0

# target #

default: unit-test build

# Test Packages

unit-test:
	go test -count=1 -v -short --cover ./...

integration-test:
	go test authentication/service_integration_test.go -count=1 -v --cover
	go test profile/service_integration_test.go -count=1 -v --cover

migration:
	go run script/run_migration/main.go
