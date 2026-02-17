.PHONY: fmt lint test build clean

fmt:
	goimports -w .

lint:
	golangci-lint run --fix

test:
	go test ./...

build:
	go build -o bin/agent-hub ./cmd/bbs
	go build -o bin/dashboard ./cmd/dashboard
	go build -o bin/client ./cmd/client

clean:
	rm -rf bin/
