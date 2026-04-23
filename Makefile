.PHONY: generate css build run test

generate:
	templ generate

css:
	tailwindcss -i web/static/css/input.css -o web/static/css/style.css --minify

build: generate css
	docker build -t conductor .

run: generate
	go run ./cmd/conductor

test:
	go test ./...
