# Project variables
NAME        := ops-tool
AUTHOR      := mattermost
URL         := https://github.com/$(AUTHOR)/$(NAME)

# Build variables
COMMIT_HASH  ?= $(shell git rev-parse --short HEAD 2>/dev/null)
BUILD_DATE   ?= $(shell date +%FT%T%z)
CUR_VERSION  ?= $(shell git describe --tags --exact-match 2>/dev/null || git describe --tags 2>/dev/null || echo "v0.0.0-$(COMMIT_HASH)")

# Go variables
LDFLAGS :="
LDFLAGS += -X github.com/$(AUTHOR)/$(NAME)/version.version=$(CUR_VERSION)
LDFLAGS += -X github.com/$(AUTHOR)/$(NAME)/version.commitHash=$(COMMIT_HASH)
LDFLAGS += -X github.com/$(AUTHOR)/$(NAME)/version.buildDate=$(BUILD_DATE)
LDFLAGS +="

## Run golangci-lint on codebase.
.PHONY: golangci-lint
golangci-lint:
	@if ! [ -x "$$(command -v golangci-lint)" ]; then \
		echo "golangci-lint is not installed. Please see https://github.com/golangci/golangci-lint#install for installation instructions."; \
		exit 1; \
	fi; \


	@echo Running golangci-lint
	golangci-lint run ./...

.PHONY: localtunnel
localtunnel:
	sudo npm install -g localtunnel
	lt --port 8080

.PHONY: run
run:
	go run ./...

.PHONY: test
test:
	go test -v -race ./...

.PHONY: build
build: clean
	@echo Building
	CGO_ENABLED=0 go build -ldflags $(LDFLAGS) -o dist/ops-tool

.PHONY: clean
clean:
	rm -rf dist/ out/

