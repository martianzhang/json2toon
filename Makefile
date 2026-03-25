.PHONY: fmt lint test test-cli test-cover build release clean

GOOS := $(shell go env GOOS)
GOPATH := $(shell go env GOPATH)
ifeq ($(GOOS),windows)
  EXE := .exe
else
  EXE :=
endif

build: fmt
	go build -o bin/j2t$(EXE) ./cmd/j2t

fmt:
	go fmt ./...

lint:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4
	$(GOPATH)/bin/golangci-lint config verify
	$(GOPATH)/bin/golangci-lint run

test: fmt
	go test ./...

cover:
	go test -coverprofile=cover.out ./...
	@echo "--- Total Coverage ---"
	@go tool cover -func=cover.out | grep -E '^total:' | grep -o '[0-9]*\.[0-9]*%' || echo "N/A"
	@echo "--- HTML report ---"
	go tool cover -html=cover.out -o cover.html

release: fmt
	GOOS=linux GOARCH=amd64 go build -o bin/j2t-linux-amd64 ./cmd/j2t
	GOOS=darwin GOARCH=amd64 go build -o bin/j2t-darwin-amd64 ./cmd/j2t
	GOOS=darwin GOARCH=arm64 go build -o bin/j2t-darwin-arm64 ./cmd/j2t
	GOOS=windows GOARCH=amd64 go build -o bin/j2t.exe ./cmd/j2t

clean:
	git clean -fdx
