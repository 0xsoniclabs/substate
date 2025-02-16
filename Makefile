.PHONY: build format test check coverage install-dev-tools

GO_BIN := $(CURDIR)/build

install-dev-tools:
	@go install golang.org/x/tools/cmd/goimports@v0.30.0
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5

check:
	@golangci-lint run --out-format=colored-line-number --timeout 5m0s

format:
	@goimports -w .
	@gofmt -s -d -w .

build:
	@go build -o $(GO_BIN)/rlp-to-protobuf ./cmd

coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

test:
	@go test -v ./...
