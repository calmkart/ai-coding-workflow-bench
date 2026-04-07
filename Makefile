.PHONY: build test clean

BINARY := workflow-bench
CMD := ./cmd/workflow-bench

build:
	go build -o $(BINARY) $(CMD)

test:
	go test ./... -count=1 -race

clean:
	rm -f $(BINARY)
	go clean -testcache
