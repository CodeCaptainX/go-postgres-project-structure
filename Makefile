.PHONY: migrate up down build

export DATABASE_URL=$(shell grep ^DATABASE_URL= .env | cut -d '=' -f2- | tr -d '"')

# Path to migrations folder
MIGRATIONS_PATH := migrations

# Target to run all migrations
db:
	goose -dir $(MIGRATIONS_PATH) create $(name) sql
  
up:
	goose -dir $(MIGRATIONS_PATH) postgres "$(DATABASE_URL)" up
# Target to drop all migrations

down:
	goose -dir $(MIGRATIONS_PATH) postgres "$(DATABASE_URL)" down-to 0
# Build Go project
build:
	go build -o bin/scan-attendance cmd/main.go
