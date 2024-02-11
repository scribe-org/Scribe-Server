.PHONY: clean build test run fmt tidy install-tools generate execute-binary

BINARY_NAME=./bin/scribe-server

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

# Format the project source code.
fmt:
	go fmt

# Sync the 'go.mod' file with dependencies in source code.
tidy:
	go mod tidy

# Install the tools for local development.
install-tools:
	go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install ariga.io/atlas/cmd/atlas@latest

# Create or update the generated source code.
generate:
	go generate -x ./...
	@$(MAKE) tidy

# Execute the binary for the project.
execute-binary:
	${BINARY_NAME}
