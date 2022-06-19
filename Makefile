ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

.PHONY: install
install:
	@go install -trimpath -ldflags="-s -w" ${ROOT_DIR}/cmd/gopack

.PHONY: lint
lint:
	@staticcheck ./...

.PHONY: test
test:
	@go test -cover -race ./...
