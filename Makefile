.PHONY: help generate build image run seed lint test clean

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-10s %s\n", $$1, $$2}'

generate: ## Run code generation
	templ generate
	tailwindcss -i web/static/css/input.css -o web/static/css/style.css --minify

build: generate ## Build local binary
	go build ./cmd/conductor

image: generate ## Build Docker image
	docker build -t conductor .

run: generate ## Run locally for development (DB at ./conductor.db)
	CONDUCTOR_DB_PATH=./conductor.db CONDUCTOR_SECURE_COOKIE=false go run ./cmd/conductor

seed: ## Seed local development database with test data (overwrites conductor.db)
	rm -f conductor.db conductor.db-shm conductor.db-wal
	CONDUCTOR_DB_PATH=./conductor.db go run ./cmd/seed

lint: generate ## Format and vet all source files (templ, Go)
	templ fmt .
	go fmt ./...
	go vet ./...
	staticcheck ./...

test: lint ## Run all tests
	go test ./...

clean: ## Remove generated files and local dev database
	rm -f web/static/css/style.css
	rm -f web/templates/*_templ.go
	rm -f conductor.db conductor.db-shm conductor.db-wal
	rm -f conductor
