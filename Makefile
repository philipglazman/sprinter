.PHONY: build

GO_PKG_LIST := $(shell go list ./...)
GO_FILES = $(shell find . -name '*.go')

.PHONY: build
build:
	go build -o build/sprinter .

.PHONY: lint
lint:
	gofmt -w ${GO_FILES} && \
	go vet ${GO_PKG_LIST} && \
	golint --set_exit_status ${GO_PKG_LIST}