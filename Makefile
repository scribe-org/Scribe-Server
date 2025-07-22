.PHONY: clean build test run fmt tidy install-tools generate generate-api generate-db execute-binary dev docs docs-serve migrate build-migrate update-data install-hooks lint

BINARY_NAME=./bin/scribe-server
MIGRATE_BINARY=./bin/migrate-scribe-data

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
	go fmt ./...

# Lint the project source code.
lint:
	@command -v revive >/dev/null 2>&1 || { echo "revive is not installed. Please run 'go install github.com/mgechev/revive@latest' first."; exit 1; }
	revive -config revive.toml ./...

# Sync the 'go.mod' file with dependencies in source code.
tidy:
	go mod tidy

# Install the tools for local development.
install-tools:
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	go install ariga.io/atlas/cmd/atlas@latest
	go install github.com/air-verse/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go get github.com/go-sql-driver/mysql
	go get github.com/glebarez/sqlite
	go get github.com/swaggo/gin-swagger
	go get github.com/swaggo/files
	@$(MAKE) install-hooks
	@echo ""
	@echo "------------------------------------------------------------"
	@echo "Go tools installed. For best results and to use these tools"
	@echo "directly from your terminal, please ensure '$(shell go env GOPATH)/bin'"
	@echo "is in your system's PATH. You might need to add:"
	@echo '  export PATH=$$(go env GOPATH)/bin:$$PATH'
	@echo "to your shell configuration file (e.g., ~/.bashrc or ~/.zshrc)"
	@echo "and then 'source' it or open a new terminal."
	@echo "------------------------------------------------------------"
	@echo ""

# Create or update the generated source code.
generate:
	@$(MAKE) generate-api
	@$(MAKE) generate-db
	@$(MAKE) docs

# Create or update the generated source code for the 'api' package.
generate-api:
	go generate -x ./api/...
	@$(MAKE) tidy

# Create or update the generated source code for the 'db' package.
generate-db:
	go generate -x ./db/...
	@$(MAKE) tidy

# Generate API documentation from code annotations.
docs:
	@echo "Generating API documentation..."
	@command -v swag >/dev/null 2>&1 || { echo "WARNING: 'swag' is not in your shell's PATH. Attempting to use absolute path. For general use, consider adding '$(shell go env GOPATH)/bin' to your PATH as instructed by 'make install-tools'."; }
	$(shell go env GOPATH)/bin/swag init --generalInfo main.go --output docs/
	@echo "Documentation generated at docs/"
	@echo ""
	@echo "After running 'make run' or 'make dev', access the documentation at:"
	@echo "  - Swagger UI: http://localhost:8080/swagger/index.html"
	@echo "  - Alternative docs: http://localhost:8080/docs/index.html"
	@echo ""

# Serve docs locally (runs server with docs available).
docs-serve: docs
	@echo "Starting server with documentation..."
	@echo "Once the server is running, you can access the API documentation at:"
	@echo "  - Swagger UI: http://localhost:8080/swagger/index.html"
	@echo "  - Alternative docs: http://localhost:8080/docs/index.html"
	@$(MAKE) run

# Execute the binary for the project.
execute-binary:
	${BINARY_NAME}

# Run the project with hot reload.
dev:
	$(shell go env GOPATH)/bin/air

# Build the migration tool.
build-migrate:
	go build -o ${MIGRATE_BINARY} ./cmd/migrate

# Run the migration tool.
migrate: build-migrate
	${MIGRATE_BINARY}

# Get data from Scribe-Data.
update-data:
	@chmod +x ./update_data.sh
	@./update_data.sh

# Install git hooks.
install-hooks:
	@mkdir -p .git/hooks
	@cp pre-commit .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
