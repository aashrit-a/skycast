BINARY := bin/skycast
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build run test lint demo gif clean

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

run:
	go run . $(ARGS)

test:
	go test ./... -race

lint:
	golangci-lint run

demo:
	go run . --demo

gif:
	vhs demo.tape

clean:
	rm -rf bin dist
