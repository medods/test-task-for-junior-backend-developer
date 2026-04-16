DEFAULT_GOAL := up

up:
	docker compose up --build

start:
	docker compose up -d --build
	docker compose logs -f

down:
	docker compose down --remove-orphans

build:
	go build ./...

test:
	go test ./... -cover

test-integration:
	@echo "Brings up postgres and migrator, then runs integration tests against localhost:5430"
	docker compose up -d postgres
	docker compose up --build migrate
	TEST_DATABASE_DSN=postgres://postgres:postgres@localhost:5430/postgres?sslmode=disable go test ./internal/repository/postgres -run TestMigrations_Staircase -v
