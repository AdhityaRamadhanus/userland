.PHONY: default test build

OS := $(shell uname)
VERSION ?= 1.0.0

# target #

default: unit-test integration-test build-api build-mail

build-api:
	@echo "Setup userland"
ifeq ($(OS),Linux)
	@echo "Build userland..."
	GOOS=linux  go build -ldflags "-s -w -X main.Version=$(VERSION)" -o api cmd/api/main.go
endif
ifeq ($(OS) ,Darwin)
	@echo "Build userland..."
	GOOS=darwin go build -ldflags "-X main.Version=$(VERSION)" -o api cmd/api/main.go
endif
	@echo "Succesfully Build for ${OS} version:= ${VERSION}"


build-mail:
	@echo "Setup userland"
ifeq ($(OS),Linux)
	@echo "Build userland..."
	GOOS=linux  go build -ldflags "-s -w -X main.Version=$(VERSION)" -o mail cmd/mailing/main.go
endif
ifeq ($(OS) ,Darwin)
	@echo "Build userland..."
	GOOS=darwin go build -ldflags "-X main.Version=$(VERSION)" -o mail cmd/mailing/main.go
endif
	@echo "Succesfully Build for ${OS} version:= ${VERSION}"

# Test Packages

unit-test:
	@go test -count=1 -v --cover ./... -tags="unit"

integration-test:
	@go test -count=1 -v --cover -tags="integration" -p 1 ./... --env-path=`pwd`/.env --config-yaml=`pwd`/config.yaml

migration:
	go run script/run_migration/main.go
