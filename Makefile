.PHONY: clean build test run

BINARY_NAME=bin/scribe-server

clean:
	go clean
	rm ${BINARY_NAME}

build:
	go build -o ${BINARY_NAME} *.go

test:
	go test ./... -v

run:
	go run .
