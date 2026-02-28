.PHONY: build run test lint sqlc migrate-up migrate-down migrate-create tidy docker-db docker-up docker-down docker-nuke docker-logs dev

build:
	go build -o bin/spend-wise-api ./cmd/api

run:
	go run ./cmd/api

test:
	go test -race -count=1 ./...

lint:
	golangci-lint run

sqlc:
	sqlc generate

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

tidy:
	go mod tidy

docker-db:
	docker compose -f ../docker-compose.yml up db pgadmin -d

docker-up:
	docker compose -f ../docker-compose.yml up -d --build

docker-down:
	docker compose -f ../docker-compose.yml down

docker-nuke:
	docker compose -f ../docker-compose.yml down -v

docker-logs:
	docker compose -f ../docker-compose.yml logs -f

DEV_DB_URL = postgres://postgres:postgres@localhost:5432/spendwise?sslmode=disable

dev:
	docker compose -f ../docker-compose.yml up db -d --wait
	migrate -path migrations -database "$(DEV_DB_URL)" up
	AUTH_BYPASS=true go run ./cmd/api
