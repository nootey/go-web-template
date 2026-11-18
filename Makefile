# Application
run:
	docker compose $(COMPOSE_OBS) $(COMPOSE_OBS_LOCAL) up -d
	go run ./cmd/api

mock:
	mockery --config=.mockery.yaml

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

# Docker
COMPOSE := -f ./docker-compose.yaml
COMPOSE_BUILD := -f ./docker-compose.build.yaml
COMPOSE_OBS := -f ./docker-compose.observability.yaml
COMPOSE_OBS_LOCAL := -f ./docker-compose.observability.local.yaml
COMPOSE_ARM := -f ./docker-compose.arm.yaml
COMPOSE_ARM_OBS := -f ./docker-compose.arm.observability.yaml

# ARM targets pull the published images by default; pass build=1 to build on the board instead
ARM_BUILD = $(if $(build),$(COMPOSE_BUILD))
ARM_FETCH = $(if $(build),--build,--pull always)

docker-up:
	docker compose $(COMPOSE_OBS) $(COMPOSE) $(COMPOSE_BUILD) up -d --build

docker-down:
	docker compose $(COMPOSE_OBS) $(COMPOSE) down

docker-restart:
	docker compose $(COMPOSE_OBS) $(COMPOSE) restart

observability-up:
	docker compose $(COMPOSE_OBS) $(COMPOSE_OBS_LOCAL) up -d

observability-down:
	docker compose $(COMPOSE_OBS) $(COMPOSE_OBS_LOCAL) down

docker-migrate:
	docker compose $(COMPOSE) $(COMPOSE_BUILD) run --rm --build migrate $(or $(type),up)

docker-seed:
	docker compose $(COMPOSE) $(COMPOSE_BUILD) run --rm --build --entrypoint seed migrate $(or $(type),core)

docker-arm-up:
	docker compose $(COMPOSE_OBS) $(COMPOSE_ARM_OBS) $(COMPOSE) $(COMPOSE_ARM) $(ARM_BUILD) up -d $(ARM_FETCH)

docker-arm-down:
	docker compose $(COMPOSE_OBS) $(COMPOSE_ARM_OBS) $(COMPOSE) $(COMPOSE_ARM) down

docker-arm-restart:
	docker compose $(COMPOSE_OBS) $(COMPOSE_ARM_OBS) $(COMPOSE) $(COMPOSE_ARM) restart

docker-arm-migrate:
	docker compose $(COMPOSE) $(COMPOSE_ARM) $(ARM_BUILD) run --rm $(ARM_FETCH) migrate $(or $(type),up)

docker-arm-seed:
	docker compose $(COMPOSE) $(COMPOSE_ARM) $(ARM_BUILD) run --rm $(ARM_FETCH) --entrypoint seed migrate $(or $(type),core)

# SQLC
sqlc:
	sqlc generate

# Tests
test:
	go test -race -count=1 -tags integration ./...

# cover=<file> writes a coverage profile
test-unit:
	go test -race -count=1 $(if $(cover),-coverpkg=./... -coverprofile=$(cover)) ./...

test-integration:
	go test -race -count=1 -tags integration -run '^TestIntegration' $(if $(cover),-coverpkg=./... -coverprofile=$(cover)) ./...

check-test-names:
	@bad=$$(git grep -l '^//go:build integration' -- '*_test.go' | xargs -r grep -HnE '^func Test' | grep -vE ':func Test(Integration|Main\()'); \
	if [ -n "$$bad" ]; then echo "$$bad"; echo "tests in integration-tagged files must start with TestIntegration"; exit 1; fi

test-coverage:
	go test -race -tags integration -coverpkg=./... -coverprofile=coverage.out ./...
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
	@echo "--- Bootstrap ---"
	mockery --config=.mockery.yaml
	@echo "--- App ---"
	@go build ./... && echo "build successful" || (echo "build failed" && exit 1)
	@golangci-lint run && echo "lint successful" || (echo "lint failed" && exit 1)
	@go test -race -count=1 -tags integration ./... && echo "tests successful" || (echo "tests failed" && exit 1)
	@echo ""
	@echo "--- Client ---"
	@cd client && pnpm run build && echo "build successful" || (echo "build failed" && exit 1)
	@cd client && pnpm run format && echo "format successful" || (echo "format failed" && exit 1)
	@cd client && pnpm run lint && echo "lint successful" || (echo "lint failed" && exit 1)