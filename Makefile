.PHONY: up down build logs ps clean cli server migrate run-cli bin-dir help mock test-server coverage lint

.DEFAULT_GOAL := help

DOCKER_COMPOSE = docker compose -f docker-compose.yml --env-file .env
DOCKER_COMPOSE_TOOL = docker compose -f docker-compose.tool.yml
DOCKER_COMPOSE_CLI = docker compose -f docker-compose.cli.yml

help:
	@echo "Usage:"
	@echo "  make up        - Start containers in foreground"
	@echo "  make up-d      - Start containers in background (detached mode)"
	@echo "  make down      - Stop and remove containers"
	@echo "  make build     - Build or rebuild all services"
	@echo "  make server    - Build only the server container"
	@echo "  make migrate   - Build only the migration container"
	@echo "  make cli       - Build the CLI tool locally"
	@echo "  make logs      - View output from containers"
	@echo "  make ps        - List running containers"
	@echo "  make clean     - Remove containers, volumes, networks, CLI binary and coverage reports"
	@echo "  make coverage  - Generate HTML test coverage reports for server and cli"
	@echo "  make help      - Display this help message"
	@echo "  make test      - Run tests for server and cli"
	@echo "  make lint      - Run linter for the server directory"

up:
	$(DOCKER_COMPOSE) up

up-d:
	$(DOCKER_COMPOSE) up -d

down:
	$(DOCKER_COMPOSE) down

build:
	$(DOCKER_COMPOSE) build

server:
	$(DOCKER_COMPOSE) build file-store

migrate:
	$(DOCKER_COMPOSE) build migrator

logs:
	$(DOCKER_COMPOSE) logs -f

ps:
	$(DOCKER_COMPOSE) ps

wire:
	$(DOCKER_COMPOSE_TOOL) run --rm wire

cli:
	cd cli/cmd && go build -o ../../fs-store .

clean:
	$(DOCKER_COMPOSE) down -v --remove-orphans
	$(DOCKER_COMPOSE_TOOL) run --rm clean

test:
	$(DOCKER_COMPOSE_TOOL) run --rm test

lint:
	$(DOCKER_COMPOSE_TOOL) run --rm lint

coverage:
	@echo "---------------------------------------"
	@echo " Generating coverage report for [Server]"
	@echo "---------------------------------------"
	cd server && go test ./... -covermode=atomic -coverprofile=coverage.out
	cd server && go tool cover -html=coverage.out -o coverage.html
	@echo "Server coverage report generated at server/coverage.html"
	@echo
	@echo "---------------------------------------"
	@echo " Generating coverage report for [CLI]"
	@echo "---------------------------------------"
	cd cli && go test ./... -covermode=atomic -coverprofile=coverage.out
	cd cli && go tool cover -html=coverage.out -o coverage.html
	@echo "CLI coverage report generated at cli/coverage.html"
