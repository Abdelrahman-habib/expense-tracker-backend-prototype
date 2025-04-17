include .env

# Build the application
all: build test

build:
	@echo "Building..."
	
	
	@go build -o main.exe cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go

# expose the application
expose:
	@zrok share reserved $(ZROK_NODE_ID) --override-endpoint http://localhost:$(PORT)

# Create DB container
docker-run:
	@docker compose up --build

# Shutdown DB container
docker-down:
	@docker compose down

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v
# Integrations Tests for the application
itest:
	@echo "Running integration tests..."
	@go test ./internal/database -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@powershell -ExecutionPolicy Bypass -Command "if (Get-Command air -ErrorAction SilentlyContinue) { \
		air; \
		Write-Output 'Watching...'; \
	} else { \
		Write-Output 'Installing air...'; \
		go install github.com/air-verse/air@latest; \
		air; \
		Write-Output 'Watching...'; \
	}"

# Generate SQLC code
sqlc:
	sqlc generate

# Run migrations
db-up:
	goose -dir internal/db/sql/migrations postgres $(DATABASE_URL) up

db-down:
	goose -dir internal/db/sql/migrations postgres $(DATABASE_URL) down

db-reset:
	goose -dir internal/db/sql/migrations postgres $(DATABASE_URL) reset

docs-private:
	@swag init -g cmd/api/main.go --ot json  --v3.1

docs-public:
	@swag init -g cmd/api/main.go -t !Webhooks --ot json  --v3.1

docs-clean:
	./scripts/clean-docs_schemas.sh --dir ./docs --files swagger.json --packages types,payloads,errors

docs-add-schema-titles:
	./scripts/add_schema_titles.sh --dir ./docs --files swagger.json

docs:
	make docs-public
	make docs-clean
	make docs-add-schema-titles


# Test setup and execution
test-setup:
	@echo "Setting up test environment..."
	@docker compose -f docker-compose.yml up postgres_test -d
	@echo "Waiting for PostgreSQL to be ready..."
	@timeout 30s sh -c 'until docker compose -f docker-compose.yml exec -T postgres_test pg_isready -U test -d testdb; do sleep 1; done'

test-teardown:
	@echo "Tearing down test environment..."
	@docker compose -f docker-compose.yml down

# Run all tests with proper setup
test-all: test-setup
	@echo "Running tests..."
	@CI=true go test ./... -v
	@make test-teardown

# Run integration tests only
test-integration: test-setup
	@echo "Running integration tests..."
	@CI=true go test ./internal/*/integration/... -v
	@make test-teardown

# Run repository tests only
test-repository: test-setup
	@echo "Running repository tests..."
	@CI=true go test ./internal/*/repository/... -v
	@make test-teardown

# Run unit tests only (no DB needed)
test-unit:
	@echo "Running unit tests..."
	@go test `go list ./... | grep -v 'integration\|repository'` -v

	
.PHONY: all sqlc db-up db-down db-reset build run test clean watch docker-run docker-down itest expose docs docs-private docs-public docs-clean docs-add-schema-titles test-setup test-teardown test-all test-integration test-repository test-unit
