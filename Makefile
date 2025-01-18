.PHONY: build format test

GO_BIN := $(CURDIR)/build

format:
	@gofmt -s -d .

build:
	@go build -o $(GO_BIN)/rlp-to-protobuf ./cmd

test:
	@go test -cover -v ./...
