NAME    := k9s
PACKAGE := github.com/k8sland/$(NAME)
VERSION := $(shell git rev-parse --short HEAD)

default: help

cover:     ## Run test coverage suite
	@go test ./... --coverprofile=cov.out
	@go tool cover --html=cov.out

osx:       ## Builds OSX CLI
	@env GOOS=darwin GOARCH=amd64 go build \
	-ldflags "-w -X ${PACKAGE}/cmd.Version=${VERSION}" \
	-a -tags netgo -o execs/${NAME} *.go


help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[38;5;69m%-30s\033[38;5;0m %s\n", $$1, $$2}'
