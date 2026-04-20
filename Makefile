SHELL := /bin/bash
DATABASE ?= postgres://postgres:postgres@localhost:5432/taskservice?sslmode=disable
.PHONY: up
up:
	@echo "--> Docker Compose Up..."
	docker compose up -d

.PHONY: down
down:
	@echo "--> Docker Compose Down..."
	docker compose down

.PHONY: build
build:
	go fmt ./...
	go vet ./...
	@echo "--> Building Docker Image..."
	docker compose build

.PHONY: migrate-up
migrate-up:
	docker run --rm --network host -v $(PWD)/migrations:/migrations migrate/migrate:v4.15.2 \
    -path=/migrations/ -database "$(DATABASE)" up

.PHONY: migrate-down
migrate-down:
	docker run --rm --network host -v $(PWD)/migrations:/migrations migrate/migrate:v4.15.2 \
    -path=/migrations/ -database "$(DATABASE)" down 1

.PHONY: migrate-create
migrate-create:
	@if [ -z "$(NAME)" ]; then \
    echo "Usage: make migrate-create NAME=migration_name"; \
    exit 1; \
	fi
	docker run --rm -v $(PWD)/migrations:/migrations migrate/migrate:v4.15.2 \
    create -ext sql -dir /migrations -seq $(NAME)

.PHONY: migrate-force
migrate-force:
	@if [ -z "$(VERSION)" ]; then \
    echo "Usage: make migrate-force VERSION=version_number"; \
    exit 1; \
	fi
	docker run --rm --network host -v $(PWD)/migrations:/migrations migrate/migrate:v4.15.2 \
    -path=/migrations/ -database "$(DATABASE)" force $(VERSION)

.PHONY: migrate-version
migrate-version:
	docker run --rm --network host -v $(PWD)/migrations:/migrations migrate/migrate:v4.15.2 \
    -path=/migrations/ -database "$(DATABASE)" version
