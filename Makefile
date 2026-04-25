.PHONY: help generate css build run seed test clean

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-10s %s\n", $$1, $$2}'

generate: ## Run templ code generation
	templ generate

css: ## Compile Tailwind CSS
	tailwindcss -i web/static/css/input.css -o web/static/css/style.css --minify

build: generate css ## Build Docker image
	docker build -t conductor .

run: generate css ## Run locally for development (DB at ./conductor.db)
	CONDUCTOR_DB_PATH=./conductor.db CONDUCTOR_SECURE_COOKIE=false go run ./cmd/conductor

seed: ## Seed local development database with test data (overwrites conductor.db)
	rm -f conductor.db conductor.db-shm conductor.db-wal
	CONDUCTOR_DB_PATH=./conductor.db go run ./cmd/seed

test: ## Run all tests
	go test ./...

clean: ## Remove generated files and local dev database
	rm -f web/static/css/style.css
	rm -f web/templates/*_templ.go
	rm -f conductor.db conductor.db-shm conductor.db-wal
