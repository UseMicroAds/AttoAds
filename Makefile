.PHONY: help start stop restart env install infra-up infra-down migrate-up migrate-down \
	api discovery verifier frontend dev dev-api dev-discovery dev-verifier dev-frontend \
	build-backend test-backend test-frontend test-contracts fmt-backend

SHELL := /bin/bash

DB_URL := postgres://microads:microads_dev@localhost:5432/microads?sslmode=disable

help:
	@echo "MicroAds developer commands"
	@echo ""
	@echo "Quick start:"
	@echo "  make start            Copy envs, start DB/Redis, run migrations"
	@echo "  make dev              Run API, workers, and frontend together"
	@echo ""
	@echo "Infrastructure:"
	@echo "  make infra-up         Start Postgres + Redis (docker compose)"
	@echo "  make infra-down       Stop Postgres + Redis"
	@echo "  make stop             Alias for infra-down"
	@echo "  make restart          Restart local infra"
	@echo ""
	@echo "Development servers:"
	@echo "  make api              Run Go API"
	@echo "  make discovery        Run discovery worker"
	@echo "  make verifier         Run verifier worker"
	@echo "  make frontend         Run Next.js frontend"
	@echo ""
	@echo "Database:"
	@echo "  make migrate-up       Run DB migrations up"
	@echo "  make migrate-down     Roll back one migration"
	@echo ""
	@echo "Tests/build:"
	@echo "  make build-backend    Build all Go binaries"
	@echo "  make test-backend     Run Go tests"
	@echo "  make test-frontend    Run frontend lint/build checks"
	@echo "  make test-contracts   Run Foundry tests"
	@echo "  make install          Install frontend deps + tidy Go deps"

start: env infra-up migrate-up
	@echo ""
	@echo "Bootstrap complete."
	@echo "Next: run 'make dev' or run 'make api' and 'make frontend' in separate terminals."

stop: infra-down

restart: infra-down infra-up

env:
	@test -f backend/.env || cp backend/.env.example backend/.env
	@test -f frontend/.env.local || cp frontend/.env.example frontend/.env.local
	@echo "Environment files are ready."

install:
	@cd backend && go mod tidy
	@cd frontend && npm install

infra-up:
	@docker compose up -d

infra-down:
	@docker compose down

migrate-up:
	@command -v migrate >/dev/null 2>&1 || (echo "Error: 'migrate' CLI not found. Install: https://github.com/golang-migrate/migrate" && exit 1)
	@migrate -path backend/migrations -database "$(DB_URL)" up

migrate-down:
	@command -v migrate >/dev/null 2>&1 || (echo "Error: 'migrate' CLI not found. Install: https://github.com/golang-migrate/migrate" && exit 1)
	@migrate -path backend/migrations -database "$(DB_URL)" down 1

api:
	@cd backend && go run ./cmd/api

discovery:
	@cd backend && go run ./cmd/discovery

verifier:
	@cd backend && go run ./cmd/verifier

frontend:
	@cd frontend && npm run dev

dev-api: api
dev-discovery: discovery
dev-verifier: verifier
dev-frontend: frontend

dev:
	@trap 'kill 0' INT TERM EXIT; \
	( cd backend && go run ./cmd/api ) & \
	( cd backend && go run ./cmd/discovery ) & \
	( cd backend && go run ./cmd/verifier ) & \
	( cd frontend && npm run dev ) & \
	wait

build-backend:
	@cd backend && go build ./...

test-backend:
	@cd backend && go test ./...

test-frontend:
	@cd frontend && npm run lint && npm run build

test-contracts:
	@cd contracts && forge test -vv

fmt-backend:
	@cd backend && go fmt ./...
