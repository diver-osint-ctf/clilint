.PHONY: test build lint fmt

test:
	go test -v

build:
	go build -o clilint

lint:
	golangci-lint run

fmt:
	go fmt ./...
