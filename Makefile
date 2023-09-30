.PHONY: clean build test run

BINARY_NAME=bin/scribe-server

# Clean any build artifacts.
clean:
	go clean
	rm ${BINARY_NAME}

# Build a binary for the project.
build:
	go build -o ${BINARY_NAME} *.go

# Run tests for the project.
test:
	go test ./... -v

# Run the project.
run:
	go run .
