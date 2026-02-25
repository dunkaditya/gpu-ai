.PHONY: build run test lint clean

BINARY=gpuctl

build:
	go build -o $(BINARY) ./cmd/gpuctl

run: build
	./$(BINARY)

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -f $(BINARY)
