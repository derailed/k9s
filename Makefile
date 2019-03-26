NAME    := k9s
PACKAGE := github.com/derailed/$(NAME)/internal
VERSION := dev
GIT     := $(shell git rev-parse --short HEAD)
DATE    := $(shell date +%FT%T%Z)

default: help

test: export GO111MODULE=on
test: export GOPROXY=https://gocenter.io
test:      ## Run all tests
	@go test ./...

cover: export GO111MODULE=on
cover: export GOPROXY=https://gocenter.io
cover:     ## Run test coverage suite
	@go test ./... --coverprofile=cov.out
	@go tool cover --html=cov.out

build: export GO111MODULE=on
build: export GOPROXY=https://gocenter.io
build:     ## Builds the CLI
	@go build \
	-ldflags "-w -X ${PACKAGE}/cmd.version=${VERSION} -X ${PACKAGE}/cmd.commit=${GIT} -X ${PACKAGE}/cmd.date=${DATE}" \
	-a -tags netgo -o execs/${NAME} *.go


help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[38;5;69m%-30s\033[38;5;38m %s\033[0m\n", $$1, $$2}'
