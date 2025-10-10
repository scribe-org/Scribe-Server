.PHONY: clean build test run fmt tidy install-tools generate generate-api generate-db execute-binary dev docs docs-serve migrate build-migrate update-data install-hooks lint

BINARY_NAME=./bin/scribe-server
MIGRATE_BINARY=./bin/migrate-scribe-data

# Default values
ENV ?= dev
GIN_MODE ?= debug

# Clean any build artifacts.
clean:
	go clean
	rm -f ${BINARY_NAME}

# Build a binary for the project.
build:
	ENV=${ENV} GIN_MODE=${GIN_MODE} go build -o ${BINARY_NAME} *.go

# Run tests for the project.
test:
	go test ./... -v

# Run the project (defaults to dev).
run: run-dev

# Run the project in development mode.
run-dev:
	ENV=dev GIN_MODE=${GIN_MODE} go run .

# Run the project in production mode.
run-prod:
	ENV=prod GIN_MODE=release go run .

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
	@echo "Go tools installed. Ensure '$(shell go env GOPATH)/bin' is in your PATH."
	@echo "Add this to ~/.bashrc or ~/.zshrc if needed:"
	@echo '  export PATH=$$(go env GOPATH)/bin:$$PATH'
	@echo "------------------------------------------------------------"
	@echo ""

# Create or update generated code and docs.
generate:
	@$(MAKE) generate-api
	@$(MAKE) generate-db
	@$(MAKE) docs

generate-api:
	go generate -x ./api/...
	@$(MAKE) tidy

generate-db:
	go generate -x ./db/...
	@$(MAKE) tidy

# Generate Swagger docs.
docs:
	@echo "Generating API documentation..."
	@command -v swag >/dev/null 2>&1 || { echo "WARNING: 'swag' not in PATH."; }
	$(shell go env GOPATH)/bin/swag init --generalInfo main.go --output docs/
	@echo "Docs generated at docs/"
	@echo "Swagger UI: http://localhost:8080/swagger/index.html"
	@echo ""

# Serve docs locally.
docs-serve: docs
	@echo "Starting server with documentation..."
	ENV=${ENV} GIN_MODE=${GIN_MODE} go run .

# Execute the binary for the project.
execute-binary:
	ENV=${ENV} GIN_MODE=${GIN_MODE} ${BINARY_NAME}

# Run the project with hot reload (dev mode).
dev:
	ENV=dev GIN_MODE=${GIN_MODE} $(shell go env GOPATH)/bin/air

# Run the project with hot reload (prod mode).
dev-prod:
	ENV=prod GIN_MODE=release $(shell go env GOPATH)/bin/air

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
