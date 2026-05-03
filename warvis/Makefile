.PHONY: build test test-unit test-integration lint clean

build:
	go build -o bin/warvis ./cmd/warvis

test-unit:
	go test ./... -short -v

test-integration:
	go test ./... -tags integration -v

test: test-unit

lint:
	go vet ./...

clean:
	rm -rf bin/
