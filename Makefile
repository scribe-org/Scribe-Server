.PHONY: clean build test run fmt tidy install-tools generate generate-api generate-db execute-binary dev

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
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install ariga.io/atlas/cmd/atlas@latest
	go install github.com/air-verse/air@latest
	go get github.com/go-sql-driver/mysql
	go get github.com/glebarez/sqlite
	@$(MAKE) install-hooks

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
