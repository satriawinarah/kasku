APP_NAME    := kasku
BINARY      := ./bin/$(APP_NAME)
MAIN_PKG    := ./cmd/kasku
TEMPL_DIR   := ./web/templates

.PHONY: all build run dev clean templ-gen templ-watch test lint fmt tools build-prod help

## Generate templ files (.templ -> _templ.go)
templ-gen:
	templ generate

## Watch and regenerate templ files on change
templ-watch:
	templ generate --watch

## Build the binary
build: templ-gen
	go build -o $(BINARY) $(MAIN_PKG)

## Run compiled binary
run: build
	$(BINARY)

## Development mode: templ watch + air live reload
## Run in two terminals: `make templ-watch` and `make dev`
dev:
	air

## Run tests
test:
	go test ./... -v -race

## Run tests with coverage
test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

## Format Go and templ files
fmt:
	gofmt -w .
	templ fmt .

## Tidy go modules
tidy:
	go mod tidy

## Install required dev tools
tools:
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/air-verse/air@latest

## Build optimized production binary
build-prod: templ-gen
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	  go build -ldflags="-s -w" -o $(BINARY)-linux-amd64 $(MAIN_PKG)

## Remove build artifacts
clean:
	rm -rf bin/ tmp/ coverage.out coverage.html

help:
	@echo "Kasku - Family Money Manager"
	@echo ""
	@echo "Available targets:"
	@grep -E '^## ' Makefile | sed 's/## /  /'
