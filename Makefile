.PHONY: up down build logs ps clean cli server run-cli bin-dir help wire mockgen test-server coverage lint build-cli-all

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
	@echo "  make cli       - Build the CLI tool locally"
	@echo "  make logs      - View output from containers"
	@echo "  make ps        - List running containers"
	@echo "  make wire      - Run wire for the server and cli"
	@echo "  make mockgen   - Run mockgen for the server and cli"
	@echo "  make clean     - Remove containers, volumes, networks, CLI binary and coverage reports"
	@echo "  make coverage  - Generate HTML test coverage reports for server and cli"
	@echo "  make help      - Display this help message"
	@echo "  make test      - Run tests for server and cli"
	@echo "  make lint      - Run linter for the server directory"
	@echo "  make build-cli-all - Build the CLI for Windows, MacOS, and Linux (amd64/arm64)"

up:
	$(DOCKER_COMPOSE) up --build

up-d:
	$(DOCKER_COMPOSE) up -d

down:
	$(DOCKER_COMPOSE) down

build:
	$(DOCKER_COMPOSE) build

server:
	$(DOCKER_COMPOSE) build file-store

logs:
	$(DOCKER_COMPOSE) logs -f

ps:
	$(DOCKER_COMPOSE) ps

wire:
	$(DOCKER_COMPOSE_TOOL) run --rm wire

mockgen:
	$(DOCKER_COMPOSE_TOOL) run --rm mockgen

cli:
	cd cli/cmd && go build -o ../../fs-store .

clean:
	$(DOCKER_COMPOSE) down -v --remove-orphans
	$(DOCKER_COMPOSE_TOOL) run --rm clean
	rm -rf bin fs-store

test:
	$(DOCKER_COMPOSE_TOOL) run --rm test

lint:
	$(DOCKER_COMPOSE_TOOL) run --rm lint

build-cli-all:
	$(DOCKER_COMPOSE_TOOL) run --rm build-cli-all
