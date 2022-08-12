# ====================================================================================
# Variables

## General Variables
APP_NAME    := ops-tool-bash-plugin

# Get current date and format like: 2022-04-27 11:32
BUILD_DATE  := $(shell date +%Y-%m-%d\ %H:%M)

## General Configuration Variables
# We don't need make's built-in rules.
MAKEFLAGS     += --no-builtin-rules
# Be pedantic about undefined variables.
MAKEFLAGS     += --warn-undefined-variables
# Set help as default target
.DEFAULT_GOAL := help

# App Code location
CONFIG_APP_CODE         += ./

## Go Variables
# Go executable
GO                           := $(shell which go)
# Extract GO version from go.mod file
GO_VERSION                   ?= $(shell grep -E '^go' go.mod | awk {'print $$2'})
# Build options
GO_BUILD_OPTS                := -mod=readonly -trimpath -buildmode=plugin
GO_TEST_OPTS                 := -mod=readonly -failfast -race
# Plugin generation dir
GO_OUT_SO_PATH               := ../../dist/plugins/bash.so

# ====================================================================================
# Colors

BLUE   := $(shell printf "\033[34m")
YELLOW := $(shell printf "\033[33m")
RED    := $(shell printf "\033[31m")
GREEN  := $(shell printf "\033[32m")
CYAN   := $(shell printf "\033[36m")
CNone  := $(shell printf "\033[0m")

# ====================================================================================
# Logger

TIME_LONG	= `date +%Y-%m-%d' '%H:%M:%S`
TIME_SHORT	= `date +%H:%M:%S`
TIME		= $(TIME_SHORT)

INFO = echo ${TIME} ${BLUE}[ .. ]${CNone}
WARN = echo ${TIME} ${YELLOW}[WARN]${CNone}
ERR  = echo ${TIME} ${RED}[FAIL]${CNone}
OK   = echo ${TIME} ${GREEN}[ OK ]${CNone}
FAIL = (echo ${TIME} ${RED}[FAIL]${CNone} && false)

# ====================================================================================
# Verbosity control hack

VERBOSE ?= 0
AT_0 := @
AT_1 :=
AT = $(AT_$(VERBOSE))

# ====================================================================================
# Targets

help: ## to get help
	@echo "Usage:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) |\
	awk 'BEGIN {FS = ":.*?## "}; {printf "make ${CYAN}%-30s${CNone} %s\n", $$1, $$2}'

.PHONY: build
build: ## to build plugin
	@$(INFO) go build 
	$(AT)$(GO) build ${GO_BUILD_OPTS} \
	-o ${GO_OUT_SO_PATH} \
	${CONFIG_APP_CODE} || ${FAIL}
	@$(OK) go build $*

.PHONY: test
test: ## to test plugin
	@$(INFO) testing...
	$(AT)$(GO) test ${GO_TEST_OPTS} ./...  || ${FAIL}
	@$(OK) testing
