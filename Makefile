.PHONY: bin fmt lint test test-cli build release clean

GOOS := $(shell go env GOOS)
GOPATH := $(shell go env GOPATH)
ifeq ($(GOOS),windows)
  EXE := .exe
else
  EXE :=
endif

build: fmt bin
	go build -o bin/j2t$(EXE) ./cmd/j2t

fmt:
	go fmt ./...

lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOPATH)/bin/golangci-lint run

test: fmt
	go test ./...

test-cli: build
	go test -v ./cmd/j2t/...

release: fmt bin
	GOOS=linux GOARCH=amd64 go build -o bin/j2t-linux-amd64 ./cmd/j2t
	GOOS=darwin GOARCH=amd64 go build -o bin/j2t-darwin-amd64 ./cmd/j2t
	GOOS=darwin GOARCH=arm64 go build -o bin/j2t-darwin-arm64 ./cmd/j2t
	GOOS=windows GOARCH=amd64 go build -o bin/j2t.exe ./cmd/j2t

clean:
	git clean -fdx
