include .env

MIGRATIONS_DIR := $(MIGRATIONS_DIR)
SWAG_ENTRY      := cmd/main.go
SWAG_OUTPUT     := docs

.PHONY: help \
        migrate-up migrate-down migrate-reset \
        migrate-status migrate-version \
        migrate-create \
        wire swag grant-admin \
        run build dev

# ── Help ──────────────────────────────────────────────────────────────────────

help:
	@echo "Available commands:"
	@echo "  make migrate-up         - Apply all up migrations"
	@echo "  make migrate-down       - Apply all down migrations"
	@echo "  make migrate-reset      - Reset the database (down then up)"
	@echo "  make migrate-status     - Show the status of all migrations"
	@echo "  make migrate-version    - Show the current migration version"
	@echo "  make migrate-create     - Create a new migration file (usage: make migrate-create name=your_migration_name)"
	@echo "  make wire               - Generate dependency injection code using Wire"
	@echo "  make swag               - Generate Swagger/OpenAPI docs"
	@echo "  make grant-admin        - Promote/create an admin (usage: make grant-admin email=user@example.com [password=secret])"
	@echo "  make run                - Run the application"
	@echo "  make build              - Build the application"
	@echo "  make dev                - Run the application in development mode"

# ── Migrations ────────────────────────────────────────────────────────────────

migrate-up:
	@echo "→ Applying up migrations..."
	@go run cmd/migrate/main.go -cmd up

migrate-down:
	@echo "→ Applying down migrations..."
	@go run cmd/migrate/main.go -cmd down

migrate-reset:
	@echo "→ Resetting database..."
	@go run cmd/migrate/main.go -cmd reset

migrate-status:
	@echo "→ Showing migration status..."
	@go run cmd/migrate/main.go -cmd status

migrate-version:
	@echo "→ Showing current migration version..."
	@go run cmd/migrate/main.go -cmd version

migrate-create:
	@if [ -z "$(name)" ]; then echo "❌  Usage: make migrate-create name=create_posts"; exit 1; fi
	@echo "→ Creating new migration: $(name)"
	@go run cmd/migrate/main.go -cmd create -name $(name)


# ── Code generation ───────────────────────────────────────────────────────────

wire:
	@echo "→ Generating dependency injection code with Wire..."
	@cd internal/wire && wire

swag:
	@echo "→ Generating Swagger/OpenAPI docs..."
	@swag init -g $(SWAG_ENTRY) -o $(SWAG_OUTPUT) --parseInternal

grant-admin:
	@if [ -z "$(email)" ]; then echo "❌  Usage: make grant-admin email=user@example.com [password=secret] "; exit 1; fi
	@echo "→ Granting admin to $(email)..."
	@go run cmd/grantadmin/main.go -email $(email) -password "$(password)"

run:
	@echo "→ Running the application..."
	@go run cmd/main.go

build:
	@echo "→ Building the application..."
	@go build -o bin/app cmd/main.go

dev:
	@air
