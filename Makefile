BINARY=ilaunch

.PHONY: build run test vet fmt

build:
	mkdir -p bin
	go build -o bin/$(BINARY) .

run:
	go run .

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w .
