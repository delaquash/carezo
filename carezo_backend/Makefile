# ==============================================================================
# Carezo Backend — Makefile
#
# Every command in this file assumes LOCAL_DB_URL, TEST_DATABASE_URL, and
# NEON_DB_URL are already exported in your shell (per ~/.bashrc). If a
# command fails with "URL cannot be empty" or similar, run:
#     source ~/.bashrc
# or re-export them manually — see the README section at the bottom of
# this file for the exact export lines.
#
# Run `make help` (or just `make`) to see every available command.
# ==============================================================================

.DEFAULT_GOAL := help

# ------------------------------------------------------------------------------
# Config — adjust here if credentials/names ever change, rather than
# hunting through every target below.
# ------------------------------------------------------------------------------
DB_CONTAINER   := carezo-postgres
APP_CONTAINER  := carezo_app
DB_USER        := carezo_user
DB_NAME        := carezo_db
RENDER_URL     := https://carezo.onrender.com
ADMIN_EMAIL    := admin@carezo.com
ADMIN_PASSWORD := Admin123!

# Neon connections MUST pin -p 5432 explicitly — some environments
# silently resolve to port 5433 otherwise (see session notes). Every
# neon-* target below already includes this.

.PHONY: help
help: ## Show this help message
	@echo "Carezo Backend — available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-28s\033[0m %s\n", $$1, $$2}'

# ==============================================================================
# BUILD & RUN
# ==============================================================================

.PHONY: run build fmt vet tidy
run: ## Run the API locally (go run)
	go run ./cmd/api

build: ## Build the production binary
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

fmt: ## Format all Go files (gofmt)
	go fmt ./...

vet: ## Run go vet across the project
	go vet ./...

tidy: ## Clean up go.mod/go.sum
	go mod tidy

# ==============================================================================
# TESTING
# ==============================================================================

.PHONY: test test-verbose test-one
test: ## Run the full test suite
	go test ./... -count=1

test-verbose: ## Run the full test suite, verbose output
	go test ./... -v -count=1

test-one: ## Run a single test by name — usage: make test-one NAME=TestRebookAfterCancellation
	go test ./internal/handlers/ -run $(NAME) -v -count=1

# ==============================================================================
# DOCKER (local dev environment)
# ==============================================================================

.PHONY: docker-up docker-down docker-rebuild docker-logs docker-psql
docker-up: ## Start local Postgres, Redis, and the app via docker-compose
	docker-compose up -d

docker-down: ## Stop all local docker-compose services
	docker-compose down

docker-rebuild: ## Force a clean rebuild of the app container (bypasses stale air/cache issues)
	docker-compose down
	docker-compose up -d --build app

docker-logs: ## Tail the app container's logs live
	docker logs $(APP_CONTAINER) --tail 100 -f

docker-psql: ## Open an interactive psql shell inside the local Postgres container
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME)

# ==============================================================================
# MIGRATIONS
#
# Every migration change gets applied to all three environments in the same
# sitting — local, test, and Neon — never just one. See session notes on
# schema drift for why this discipline matters.
# ==============================================================================

.PHONY: migrate-create migrate-up-local migrate-up-test migrate-up-neon migrate-up-all
.PHONY: migrate-version-local migrate-version-test migrate-version-neon migrate-version-all
.PHONY: migrate-down-local migrate-force-local migrate-force-neon

migrate-create: ## Create a new migration pair — usage: make migrate-create NAME=add_push_tokens_table
	migrate create -ext sql -dir ./migrations -seq $(NAME)

migrate-up-local: ## Apply all pending migrations to LOCAL
	migrate -path ./migrations -database "$(LOCAL_DB_URL)" up

migrate-up-test: ## Apply all pending migrations to TEST db
	migrate -path ./migrations -database "$(TEST_DATABASE_URL)" up

migrate-up-neon: ## Apply all pending migrations to NEON (production)
	migrate -path ./migrations -database "$(NEON_DB_URL)" up

migrate-up-all: migrate-up-local migrate-up-test migrate-up-neon ## Apply migrations to ALL THREE environments in one go

migrate-version-local: ## Show current migration version — LOCAL
	migrate -path ./migrations -database "$(LOCAL_DB_URL)" version

migrate-version-test: ## Show current migration version — TEST
	migrate -path ./migrations -database "$(TEST_DATABASE_URL)" version

migrate-version-neon: ## Show current migration version — NEON
	migrate -path ./migrations -database "$(NEON_DB_URL)" version

migrate-version-all: ## Show migration version across ALL THREE environments — run this first when debugging drift
	@echo "--- LOCAL ---"; migrate -path ./migrations -database "$(LOCAL_DB_URL)" version
	@echo "--- TEST ---"; migrate -path ./migrations -database "$(TEST_DATABASE_URL)" version
	@echo "--- NEON ---"; migrate -path ./migrations -database "$(NEON_DB_URL)" version

migrate-down-local: ## Roll back ONE migration on LOCAL only (never run against Neon casually)
	migrate -path ./migrations -database "$(LOCAL_DB_URL)" down 1

migrate-force-local: ## Clear a dirty migration state on LOCAL — usage: make migrate-force-local V=9
	migrate -path ./migrations -database "$(LOCAL_DB_URL)" force $(V)

migrate-force-neon: ## Clear a dirty migration state on NEON — usage: make migrate-force-neon V=9
	migrate -path ./migrations -database "$(NEON_DB_URL)" force $(V)

# ==============================================================================
# DATABASE INSPECTION — the checks that ended up mattering most this session
# ==============================================================================

.PHONY: db-columns-local db-columns-neon db-constraints-local db-constraints-neon
.PHONY: db-shell-local db-shell-neon db-diff-columns

db-columns-local: ## List all columns for a table — usage: make db-columns-local TABLE=drivers
	psql "$(LOCAL_DB_URL)" -c "SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = '$(TABLE)' ORDER BY ordinal_position;"

db-columns-neon: ## List all columns for a table on NEON — usage: make db-columns-neon TABLE=drivers
	psql "$(NEON_DB_URL)" -p 5432 -c "SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = '$(TABLE)' ORDER BY ordinal_position;"

db-constraints-local: ## List all CHECK constraints on a table — usage: make db-constraints-local TABLE=users
	psql "$(LOCAL_DB_URL)" -c "SELECT conname, pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid = '$(TABLE)'::regclass AND contype = 'c';"

db-constraints-neon: ## Same, against NEON
	psql "$(NEON_DB_URL)" -p 5432 -c "SELECT conname, pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid = '$(TABLE)'::regclass AND contype = 'c';"

db-diff-columns: ## Diff a table's columns between LOCAL and NEON — usage: make db-diff-columns TABLE=bookings
	diff \
		<(psql "$(LOCAL_DB_URL)" -tAc "SELECT column_name FROM information_schema.columns WHERE table_name='$(TABLE)' ORDER BY column_name") \
		<(psql "$(NEON_DB_URL)" -p 5432 -tAc "SELECT column_name FROM information_schema.columns WHERE table_name='$(TABLE)' ORDER BY column_name") \
		&& echo "No drift — tables match." || echo "^ Drift found — left is LOCAL, right is NEON"

db-shell-local: ## Open a raw psql shell against LOCAL
	psql "$(LOCAL_DB_URL)"

db-shell-neon: ## Open a raw psql shell against NEON
	psql "$(NEON_DB_URL)" -p 5432

# ==============================================================================
# DESTRUCTIVE — LOCAL ONLY. Never point these at Neon.
# ==============================================================================

.PHONY: db-reset-local
db-reset-local: ## DESTRUCTIVE: terminate connections, drop, and recreate the LOCAL database (fixes dirty migration state)
	@echo "This will DROP and recreate $(DB_NAME) on LOCAL. Ctrl+C now to abort."
	@sleep 3
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d postgres -c \
		"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$(DB_NAME)' AND pid <> pg_backend_pid();"
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d postgres -c "DROP DATABASE $(DB_NAME);"
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d postgres -c "CREATE DATABASE $(DB_NAME);"
	$(MAKE) migrate-up-local

# ==============================================================================
# ADMIN / LIVE API HELPERS — against the deployed Render instance
# ==============================================================================

.PHONY: admin-token health cars-check popular-check nearby-check

admin-token: ## Fetch a fresh admin access token from the LIVE Render API and print it
	@curl -s -X POST "$(RENDER_URL)/api/auth/login" \
		-H "Content-Type: application/json" \
		-d '{"email":"$(ADMIN_EMAIL)","password":"$(ADMIN_PASSWORD)"}' \
		| jq -r '.data.access_token'

health: ## Hit the live Render health check endpoint
	curl -s "$(RENDER_URL)/health" | jq '.'

cars-check: ## Sanity check: count + first few cars from the live API
	curl -s "$(RENDER_URL)/api/cars?per_page=20" | jq '.data.cars | length as $$n | "\($$n) cars found"'

popular-check: ## Sanity check: does /api/cars/popular return real data (not null/duplicated JSON)?
	curl -s "$(RENDER_URL)/api/cars/popular" | jq '.'

nearby-check: ## Sanity check: nearby cars for a given city — usage: make nearby-check CITY=Lagos
	curl -s "$(RENDER_URL)/api/cars/nearby?city=$(CITY)" | jq '.'