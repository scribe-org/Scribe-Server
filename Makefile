.PHONY: clean build test run fmt tidy install-tools generate generate-api generate-db execute-binary dev

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
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install ariga.io/atlas/cmd/atlas@latest
	go install github.com/air-verse/air@latest

# Create or update the generated source code.
generate:
	@$(MAKE) generate-api
	@$(MAKE) generate-db

# Create or update the generated source code for the 'api' package.
generate-api:
	go generate -x ./api/...
	@$(MAKE) tidy

# Create or update the generated source code for the 'db' package.
generate-db:
	go generate -x ./db/...
	@$(MAKE) tidy

# Execute the binary for the project.
execute-binary:
	${BINARY_NAME}

# Run the project with hot reload.
dev:
	$(shell go env GOPATH)/bin/air
