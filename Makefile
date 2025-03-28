PROJECT := $(shell basename "$(PWD)")
GO_BIN := $(CURDIR)/build

# compile time variables will be injected into the app
APP_VERSION := 1.0
BUILD_DATE := $(shell date "+%a, %d %b %Y %T")
BUILD_COMPILER := $(shell go version)
BUILD_COMMIT := $(shell git show --format="%H" --no-patch)
BUILD_COMMIT_TIME := $(shell git show --format="%cD" --no-patch)
GOPROXY ?= "https://proxy.golang.org,direct"

.PHONY: all clean help test

all: compare-substate rlp-to-protobuf

compare-substate:
	GOPROXY=$(GOPROXY) \
	go build -ldflags "-s -w" \
	-o $(GO_BIN)/compare-substate \
	./cmd/compare-substate

rlp-to-protobuf:
	GOPROXY=$(GOPROXY) \
	go build -ldflags "-s -w" \
	-o $(GO_BIN)/rlp-to-protobuf \
	./cmd/rlp-to-protobuf

test:
	@go test ./...

clean:
	rm -fr ./build/*

help: Makefile
	@echo "Choose a make command in "$(PROJECT)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

format:
	@goimports -w ./
	@gofmt -s -d -w ./

check:
	@golangci-lint run -c .golangci.yml ./...
