.PHONY: help dev-api dev-web build build-web build-api test test-web lint up down logs backup-db restore-db clean

COMPOSE = docker compose --env-file .env -f docker-compose.yml

help:
	@echo "Available commands:"
	@echo "  make up          - Build and start all services with Docker Compose (needs .env, see .env.example)"
	@echo "  make down        - Stop and remove the services (the database volume is kept)"
	@echo "  make logs        - Follow the service logs"
	@echo "  make dev-api     - Run the Go API locally (APP_ENV=development, SQLite in apps/api/data)"
	@echo "  make dev-web     - Run the SvelteKit dev server (proxies /api to the local API)"
	@echo "  make build       - Build the Go API binary and the static web output"
	@echo "  make test        - Run the Go tests"
	@echo "  make lint        - gofmt / go vet / svelte-check"
	@echo "  make backup-db   - Create a safe online SQLite backup"
	@echo "  make restore-db  - Restore a backup: make restore-db FILE=path/to/backup.db.gz"
	@echo "  make clean       - Remove build outputs"

# Fail early with a clear message instead of a confusing compose error.
.env:
	@echo "Missing .env - run: cp .env.example .env  (then set ADMIN_PASSWORD and SESSION_SECRET)"; exit 1

up: .env
	$(COMPOSE) up -d --build

down: .env
	$(COMPOSE) down

logs: .env
	$(COMPOSE) logs -f --tail=100

dev-api:
	cd apps/api && APP_ENV=development DATABASE_PATH=./data/cards.db MIGRATIONS_PATH=./migrations go run ./cmd/server

dev-web:
	cd apps/web && npm run dev

build: build-web build-api

build-web:
	cd apps/web && npm ci && npm run build

build-api:
	cd apps/api && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/cards-api ./cmd/server

test:
	cd apps/api && go test ./...

lint:
	@cd apps/api && test -z "$$(gofmt -l .)" || (echo "gofmt needed on:"; gofmt -l .; exit 1)
	cd apps/api && go vet ./...
	cd apps/web && npm run check && npm run check:i18n

backup-db:
	./scripts/backup-sqlite.sh

restore-db:
	@if [ -z "$(FILE)" ]; then echo "Usage: make restore-db FILE=path/to/backup.db.gz"; exit 1; fi
	./scripts/restore-sqlite.sh $(FILE)

clean:
	rm -rf apps/api/bin apps/web/build apps/web/.svelte-kit
