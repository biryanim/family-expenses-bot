GOLANGCI_LINT_VERSION ?= v2.13.2
GOLANGCI_LINT ?= golangci-lint
GOOSE_VERSION ?= v3.29.0


.PHONY: help build test vet lint lint-fix fmt tidy check generate generate-web run up down logs lint-install llm-check migrate-up migrate-down migrate-reset migrate-status migrate-create

build:
	go build ./...

test:
	go test ./...

lint:
	$(GOLANGCI_LINT) run ./...

lint-fix:
	$(GOLANGCI_LINT) run --fix ./...

check: lint test

generate:
	go generate ./...

generate-web:
	npm --prefix web run generate:api

run:
	@test -f .env || (echo ".env not found: cp .env.example .env" && exit 1)
	set -a && . ./.env && set +a && go run ./cmd/app

llm-check:
	@test -f .env || (echo ".env not found: cp .env.example .env" && exit 1)
	set -a && . ./.env && set +a && go run ./cmd/llmcheck

up:
	docker compose up --build

down:
	docker compose down

migrate-up:
	docker compose run --rm migrate up

migrate-down:
	docker compose run --rm migrate down

migrate-reset:
	docker compose run --rm migrate reset

migrate-status:
	docker compose run --rm migrate status

migrate-create:
	@test -n "$(name)" || (echo "usage: make migrate-create name=<migration_name>" && exit 1)
	go run github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION) -dir migrations -s create $(name) sql


lint-install:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
