.PHONY: build run run-debug provision test lint clean

BINARY=/var/tmp/gpu-ai/gpuctl

build:
	@mkdir -p /var/tmp/gpu-ai
	go build -o $(BINARY) ./cmd/gpuctl

run: build
	$(BINARY)

run-debug: build
	$(BINARY) serve --log-level=debug --log-format=text

provision: build
	$(BINARY) provision $(ARGS)

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -f $(BINARY)
