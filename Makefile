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
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Lint
lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix