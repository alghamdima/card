.PHONY: help dev build test lint up down migrate-up backup-db restore-db clean

help:
	@echo "Available commands:"
	@echo "  make up          - Build and start all Docker services via Docker Compose"
	@echo "  make down        - Stop and tear down Docker services"
	@echo "  make build       - Build both Go API and SvelteKit web static output"
	@echo "  make test        - Run tests for Go API"
	@echo "  make backup-db   - Create a safe online SQLite backup"
	@echo "  make clean       - Clean temporary build outputs and node_modules"

up: build
	docker compose -f infra/docker-compose.yml up -d --build

down:
	docker compose -f infra/docker-compose.yml down

build: build-web build-api

build-web:
	cd apps/web && npm install && npm run build

build-api:
	cd apps/api && go mod tidy && CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/cards-api ./cmd/server/main.go

test:
	cd apps/api && go test -v ./...

backup-db:
	./scripts/backup-sqlite.sh

restore-db:
	@if [ -z "$(FILE)" ]; then echo "Usage: make restore-db FILE=path/to/backup.db.gz"; exit 1; fi
	./scripts/restore-sqlite.sh $(FILE)

clean:
	rm -rf apps/api/bin apps/web/build apps/web/.svelte-kit
