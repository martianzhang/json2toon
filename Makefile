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
	@echo "Running linter..."
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4
	$(GOPATH)/bin/golangci-lint config verify
	$(GOPATH)/bin/golangci-lint run

test: fmt
	go test ./...

cover:
	@go clean -testcache
	go test -coverprofile="cover.out" ./...
	@echo --- Total Coverage ---
	@go tool cover -func="cover.out" | grep "total:" | grep -o "[0-9]*\.[0-9]*.*"
	@echo --- HTML report ---
	go tool cover -html="cover.out" -o "cover.html"

release: fmt
	@echo "Building for multiple platforms..."
	@echo "Linux..."
	@-go env -w GOOS=linux GOARCH=amd64 && go build -o bin/j2t-linux-amd64 ./cmd/j2t
	@-go env -w GOOS=linux GOARCH=arm64 && go build -o bin/j2t-linux-arm64 ./cmd/j2t
	@echo "Darwin..."
	@-go env -w GOOS=darwin GOARCH=amd64 && go build -o bin/j2t-darwin-amd64 ./cmd/j2t
	@-go env -w GOOS=darwin GOARCH=arm64 && go build -o bin/j2t-darwin-arm64 ./cmd/j2t
	@echo "Windows..."
	@-go env -w GOOS=windows GOARCH=amd64 && go build -o bin/j2t.exe ./cmd/j2t
	@-go env -w GOOS=windows GOARCH=arm64 && go build -o bin/j2t-arm64.exe ./cmd/j2t
	@go env -u GOOS GOARCH

clean:
	git clean -fdx
