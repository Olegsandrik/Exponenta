include .env
export $(shell sed 's/=.*//' .env)

run-prod:
	docker-compose -f Docker-compose.yml up -d postgres elasticsearch
	docker build -t exponent-image .
	sleep 1
	docker-compose -f Docker-compose.yml up -d myAPI


stop-prod:
	docker-compose -f Docker-compose.yml down

linter:
	golangci-lint run --config=golangci.yaml

setup:
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.57.1

migrate-up:
	goose -dir db/migrations postgres "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWD}@${POSTGRES_MIGRATION_HOST}:${POSTGRES_PORT}/${POSTGRES_DB_NAME}?sslmode=disable" up

migrate-down:
	goose -dir db/migrations postgres "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWD}@${POSTGRES_MIGRATION_HOST}:${POSTGRES_PORT}/${POSTGRES_DB_NAME}?sslmode=disable" down