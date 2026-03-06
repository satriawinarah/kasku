APP_NAME    := kasku
BINARY      := ./bin/$(APP_NAME)
MAIN_PKG    := ./cmd/kasku
TEMPL_DIR   := ./web/templates

# Deploy settings
DEPLOY_DIR  := /opt/kasku
SERVICE     := kasku

.PHONY: all build run dev clean templ-gen templ-watch test lint fmt tools build-prod help \
        deploy deploy-setup deploy-stop deploy-status deploy-logs deploy-uninstall deploy-nginx

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

## ------ Deployment (run on Ubuntu server) ------

## First-time server setup: install Go, templ, create system user and directories
deploy-setup:
	@echo "==> Installing dependencies..."
	@if ! command -v go >/dev/null 2>&1; then \
		echo "Installing Go..."; \
		sudo apt-get update && sudo apt-get install -y golang-go; \
	else \
		echo "Go already installed: $$(go version)"; \
	fi
	@if ! command -v templ >/dev/null 2>&1; then \
		echo "Installing templ..."; \
		go install github.com/a-h/templ/cmd/templ@latest; \
	else \
		echo "templ already installed: $$(templ version)"; \
	fi
	@echo "==> Creating system user and directories..."
	@sudo id -u $(SERVICE) >/dev/null 2>&1 || sudo useradd --system --no-create-home --shell /usr/sbin/nologin $(SERVICE)
	@sudo mkdir -p $(DEPLOY_DIR)
	@sudo chown $(SERVICE):$(SERVICE) $(DEPLOY_DIR)
	@echo "==> Setup complete."

## Build, install binary, and start/restart the systemd daemon
deploy: templ-gen
	@echo "==> Building production binary..."
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINARY) $(MAIN_PKG)
	@echo "==> Installing to $(DEPLOY_DIR)..."
	@sudo cp $(BINARY) $(DEPLOY_DIR)/$(APP_NAME)
	@sudo chmod 755 $(DEPLOY_DIR)/$(APP_NAME)
	@if [ ! -f $(DEPLOY_DIR)/.env ]; then \
		echo "==> Generating .env with random session secret..."; \
		SECRET=$$(openssl rand -hex 32); \
		printf 'PORT=8080\nDB_PATH=$(DEPLOY_DIR)/kasku.db\nSESSION_SECRET=%s\nAPP_URL=http://localhost:8080\n' "$$SECRET" | sudo tee $(DEPLOY_DIR)/.env > /dev/null; \
		sudo chown $(SERVICE):$(SERVICE) $(DEPLOY_DIR)/.env; \
		sudo chmod 600 $(DEPLOY_DIR)/.env; \
		echo "    Created $(DEPLOY_DIR)/.env — edit APP_URL to match your server address"; \
	else \
		echo "    $(DEPLOY_DIR)/.env already exists, keeping it"; \
	fi
	@echo "==> Installing systemd service..."
	@sudo cp deploy/kasku.service /etc/systemd/system/$(SERVICE).service
	@sudo systemctl daemon-reload
	@sudo systemctl enable $(SERVICE)
	@sudo systemctl restart $(SERVICE)
	@echo ""
	@echo "==> Deploy complete! Checking status..."
	@sleep 1
	@sudo systemctl status $(SERVICE) --no-pager || true

## Stop the kasku daemon
deploy-stop:
	@sudo systemctl stop $(SERVICE)
	@echo "Stopped $(SERVICE)"

## Show daemon status
deploy-status:
	@sudo systemctl status $(SERVICE) --no-pager

## Tail daemon logs (Ctrl+C to exit)
deploy-logs:
	@sudo journalctl -u $(SERVICE) -f

## Completely uninstall: stop service, remove files
deploy-uninstall:
	@echo "==> Stopping and disabling service..."
	@sudo systemctl stop $(SERVICE) 2>/dev/null || true
	@sudo systemctl disable $(SERVICE) 2>/dev/null || true
	@sudo rm -f /etc/systemd/system/$(SERVICE).service
	@sudo systemctl daemon-reload
	@echo "==> Removing $(DEPLOY_DIR) (keeping database backup)..."
	@if [ -f $(DEPLOY_DIR)/kasku.db ]; then \
		BACKUP="/tmp/kasku-db-backup-$$(date +%Y%m%d%H%M%S).db"; \
		sudo cp $(DEPLOY_DIR)/kasku.db $$BACKUP; \
		echo "    Database backed up to $$BACKUP"; \
	fi
	@sudo rm -rf $(DEPLOY_DIR)
	@sudo userdel $(SERVICE) 2>/dev/null || true
	@echo "==> Uninstall complete."

## Install Nginx reverse proxy config
deploy-nginx:
	@echo "==> Installing Nginx config..."
	@if ! command -v nginx >/dev/null 2>&1; then \
		echo "Installing Nginx..."; \
		sudo apt-get update && sudo apt-get install -y nginx; \
	fi
	@sudo cp deploy/nginx.conf /etc/nginx/sites-available/$(SERVICE)
	@sudo ln -sf /etc/nginx/sites-available/$(SERVICE) /etc/nginx/sites-enabled/$(SERVICE)
	@sudo nginx -t && sudo systemctl reload nginx
	@echo "==> Nginx configured. Access the app at http://202.10.42.99"

help:
	@echo "Kasku - Family Money Manager"
	@echo ""
	@echo "Development:"
	@echo "  templ-gen      Generate templ files"
	@echo "  templ-watch    Watch and regenerate templ files"
	@echo "  build          Build the binary"
	@echo "  run            Build and run"
	@echo "  dev            Hot reload with air"
	@echo "  test           Run tests"
	@echo "  fmt            Format code"
	@echo "  tidy           Run go mod tidy"
	@echo "  tools          Install dev tools (templ, air)"
	@echo "  clean          Remove build artifacts"
	@echo ""
	@echo "Deployment (run on Ubuntu server):"
	@echo "  deploy-setup     First-time setup (Go, templ, system user)"
	@echo "  deploy           Build + install + restart daemon"
	@echo "  deploy-stop      Stop the daemon"
	@echo "  deploy-status    Show daemon status"
	@echo "  deploy-logs      Tail daemon logs"
	@echo "  deploy-nginx     Install Nginx reverse proxy"
	@echo "  deploy-uninstall Remove everything (backs up DB)"
