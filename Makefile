.PHONY: help generate build dist image run seed lint test clean

DIST_DIR := dist

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-10s %s\n", $$1, $$2}'

generate: ## Run code generation
	templ generate
	tailwindcss -i web/static/css/input.css -o web/static/css/style.css --minify

build: generate ## Build local binary
	go build ./cmd/conductor

dist: generate ## Cross-compile release binaries for all platforms (output: dist/)
	@mkdir -p $(DIST_DIR)
	GOOS=darwin  GOARCH=amd64       CGO_ENABLED=0 go build -ldflags="-s -w" -o $(DIST_DIR)/conductor-darwin-amd64  ./cmd/conductor
	GOOS=darwin  GOARCH=arm64       CGO_ENABLED=0 go build -ldflags="-s -w" -o $(DIST_DIR)/conductor-darwin-arm64  ./cmd/conductor
	GOOS=linux   GOARCH=amd64       CGO_ENABLED=0 go build -ldflags="-s -w" -o $(DIST_DIR)/conductor-linux-amd64   ./cmd/conductor
	GOOS=linux   GOARCH=arm64       CGO_ENABLED=0 go build -ldflags="-s -w" -o $(DIST_DIR)/conductor-linux-arm64   ./cmd/conductor
	GOOS=linux   GOARCH=arm  GOARM=7 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(DIST_DIR)/conductor-linux-armhf  ./cmd/conductor

image: generate ## Build Docker image for all platforms
	docker build --platform linux/amd64,linux/arm64 -t conductor .

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
	rm -rf $(DIST_DIR)
