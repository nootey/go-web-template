# Application
run:
	go run cmd/api/main.go

# Migrations (using Go wrapper)
migrate-up:
	go run cmd/migrate/main.go up

migrate-down:
	go run cmd/migrate/main.go down

migrate-status:
	go run cmd/migrate/main.go status

migrate-create:
	go run cmd/migrate/main.go create $(NAME)

migrate-reset:
	go run cmd/migrate/main.go reset

migrate-fresh:
	go run cmd/migrate/main.go reset
	go run cmd/migrate/main.go up
	go run cmd/seed/main.go $(or $(SEED),core)

# Seeding
seed:
	go run cmd/seed/main.go core

seed-full:
	go run cmd/seed/main.go full

# SQLC
sqlc:
	sqlc generate

# Tests
test:
	go test -race -count=1 ./...

test-coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Lint
lint:
	golangci-lint run

lint-fix:
	gofmt -w .
	golangci-lint run --fix

tidy:
	go mod tidy
	go mod verify

# Pre push checklist
pre-push:
	@echo "--- App ---"
	@go build ./... && echo "build successful" || (echo "build failed" && exit 1)
	@golangci-lint run && echo "lint successful" || (echo "lint failed" && exit 1)
	@go test -race -count=1 ./... && echo "tests successful" || (echo "tests failed" && exit 1)
	@echo ""
	@echo "--- Client ---"
	@cd client && pnpm run build && echo "build successful" || (echo "build failed" && exit 1)
	@cd client && pnpm run format && echo "format successful" || (echo "format failed" && exit 1)
	@cd client && pnpm run lint && echo "lint successful" || (echo "lint failed" && exit 1)