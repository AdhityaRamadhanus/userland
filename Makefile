.PHONY: default test build

OS := $(shell uname)
VERSION ?= 1.0.0

# target #

default: unit-test build

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
	go test -count=1 -v --cover ./... -tags="unit"

service-integration-test:
	@go test ./... -count=1 -v --cover -tags="authentication" | { grep -v 'no test files'; true; }
	@go test ./... -count=1 -v --cover -tags="profile" | { grep -v 'no test files'; true; }

repository-integration-test:
	@go test ./... -count=1 -v --cover -tags="redis_repository" | { grep -v 'no test files'; true; }
	@go test ./... -count=1 -v --cover -tags="postgres_repository" | { grep -v 'no test files'; true; }

integration-test: service-integration-test repository-integration-test

migration:
	go run script/run_migration/main.go
