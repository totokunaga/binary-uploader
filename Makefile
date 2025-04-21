.PHONY: up down build logs ps clean cli server migrate run-cli bin-dir help

.DEFAULT_GOAL := help

DOCKER_COMPOSE = docker compose -f docker-compose.yml --env-file .env

help:
	@echo "Usage:"
	@echo "  make up        - Start containers in foreground"
	@echo "  make up-d      - Start containers in background (detached mode)"
	@echo "  make down      - Stop and remove containers"
	@echo "  make build     - Build or rebuild all services"
	@echo "  make server    - Build only the server container"
	@echo "  make migrate   - Build only the migration container"
	@echo "  make cli       - Build the CLI tool locally"
	@echo "  make run-cli   - Run the CLI tool (builds it first if needed)"
	@echo "  make logs      - View output from containers"
	@echo "  make ps        - List running containers"
	@echo "  make clean     - Remove containers, volumes, and networks"
	@echo "  make help      - Display this help message"

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
	pushd server/cmd && wire && popd
	pushd cli/cmd && wire && popd

bin-dir:
	@mkdir -p bin

cli: bin-dir
	@echo "Building CLI tool locally..."
	cd cli && go build -o ../bin/cli-tool ./cmd

run-cli: cli
	@echo "Running CLI tool..."
	@./bin/cli-tool

clean:
	$(DOCKER_COMPOSE) down -v --remove-orphans
	@echo "Containers, volumes, and networks removed"
	@rm -f bin/cli-tool 2>/dev/null || true
	@echo "CLI binary removed" 