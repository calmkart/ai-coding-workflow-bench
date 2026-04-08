.PHONY: build test lint fmt vet clean install

BINARY := workflow-bench
CMD := ./cmd/workflow-bench
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -X main.version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(CMD)

install:
	go install -ldflags "$(LDFLAGS)" $(CMD)

test:
	go test ./... -count=1 -race

lint:
	@command -v staticcheck >/dev/null 2>&1 || { echo "install: go install honnef.co/go/tools/cmd/staticcheck@latest"; exit 1; }
	staticcheck ./...

fmt:
	@test -z "$$(gofmt -l cmd/ internal/)" || { gofmt -l cmd/ internal/ && echo "Run 'gofmt -w cmd/ internal/' to fix"; exit 1; }

vet:
	go vet ./...

coverage:
	go test ./... -count=1 -coverprofile=coverage.out
	go tool cover -func=coverage.out

clean:
	rm -f $(BINARY) coverage.out
	go clean -testcache

check: fmt vet lint test
