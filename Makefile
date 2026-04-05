.PHONY: all build clean run watch test lint coverage fmt setup \
        db/up db/down db/destroy db/psql \
        image chart up down help

# ---------------------------------------------------------------------------
# Variables
# ---------------------------------------------------------------------------

IMAGE_NAME  = go-service-starter
TAG         = latest
NAMESPACE   = default
APP_NAME    = go-service
BINARY      = out/main
GO_FLAGS    = -race

# Load .env if it exists so that `make run` / `make watch` pick up local vars.
# Variables already set in the shell environment take precedence.
-include .env
export

# Detect kind for image push
HAS_KIND := $(shell command -v kind > /dev/null 2>&1 && echo "yes" || echo "no")
# Detect air for live-reload
HAS_AIR  := $(shell command -v air > /dev/null 2>&1 && echo "yes" || echo "no")

# ---------------------------------------------------------------------------
# Default
# ---------------------------------------------------------------------------

all: build

# ---------------------------------------------------------------------------
# Build
# ---------------------------------------------------------------------------

build:
	@mkdir -p out/
	@go mod download
	@go build -o $(BINARY) .
	@echo "--------------------------------------"
	@echo ">   OK: Build done → $(BINARY)"
	@echo "--------------------------------------"

# ---------------------------------------------------------------------------
# Run (local, with env vars from .env)
# ---------------------------------------------------------------------------

## Run the server locally. Reads DATABASE_URL from .env or the current shell
## environment. Start Postgres first with `make db/up`.
run: build
	@if [ -z "$$DATABASE_URL" ]; then \
		echo ">   ERROR: DATABASE_URL is not set."; \
		echo ">   Copy .env.example to .env and fill in values, or export the variable."; \
		exit 1; \
	fi
	@echo "--------------------------------------"
	@echo ">   Starting server on port $${PORT:-8080}..."
	@echo "--------------------------------------"
	./$(BINARY)

## Watch for source changes and live-reload (requires `make setup` first).
watch:
	@if [ "$(HAS_AIR)" = "no" ]; then \
		echo ">   ERROR: air is not installed. Run: make setup"; \
		exit 1; \
	fi
	@if [ -z "$$DATABASE_URL" ]; then \
		echo ">   ERROR: DATABASE_URL is not set. Copy .env.example → .env"; \
		exit 1; \
	fi
	air

# ---------------------------------------------------------------------------
# Test
# ---------------------------------------------------------------------------

test:
	@go test $(GO_FLAGS) ./...
	@echo "--------------------------------------"
	@echo ">   OK: Tests passed."
	@echo "--------------------------------------"

## Run tests and print a per-package coverage summary.
coverage:
	@mkdir -p out/
	@go test $(GO_FLAGS) -coverprofile=out/coverage.out -covermode=atomic ./...
	@echo ""
	@echo "--------------------------------------"
	@echo ">   Coverage summary"
	@echo "--------------------------------------"
	@go tool cover -func=out/coverage.out | tail -1
	@echo ""
	@echo ">   Full breakdown:"
	@go tool cover -func=out/coverage.out
	@echo ""
	@echo ">   HTML report: out/coverage.html"
	@go tool cover -html=out/coverage.out -o out/coverage.html
	@echo "--------------------------------------"

# ---------------------------------------------------------------------------
# Format
# ---------------------------------------------------------------------------

## Auto-format all Go source files.
fmt:
	@gofmt -w .
	@echo "--------------------------------------"
	@echo ">   OK: Format done."
	@echo "--------------------------------------"

# ---------------------------------------------------------------------------
# Lint
# ---------------------------------------------------------------------------

## Run go vet. If golangci-lint is installed (make setup), it runs that too.
lint:
	@go vet ./...
	@if command -v golangci-lint > /dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo ">   NOTE: golangci-lint not found. Run 'make setup' for deeper linting."; \
	fi
	@echo "--------------------------------------"
	@echo ">   OK: Lint passed."
	@echo "--------------------------------------"

# ---------------------------------------------------------------------------
# Clean
# ---------------------------------------------------------------------------

clean:
	@rm -rf out/
	@docker rmi -f $(IMAGE_NAME):$(TAG) 2>/dev/null || true
	@echo "--------------------------------------"
	@echo ">   OK: Clean done."
	@echo "--------------------------------------"

# ---------------------------------------------------------------------------
# Local database (docker compose)
# ---------------------------------------------------------------------------

## Start Postgres in the background.
db/up:
	@docker compose up -d postgres
	@echo "--------------------------------------"
	@echo ">   Waiting for Postgres to be ready..."
	@echo "--------------------------------------"
	@docker compose exec postgres sh -c \
		'until pg_isready -U starter -d starter; do sleep 1; done'
	@echo ">   OK: Postgres is ready."
	@echo "--------------------------------------"

## Stop and remove the Postgres container (data volume is preserved).
db/down:
	@docker compose down
	@echo "--------------------------------------"
	@echo ">   OK: Postgres stopped."
	@echo "--------------------------------------"

## Stop Postgres and destroy all data (removes the volume).
db/destroy:
	@docker compose down -v
	@echo "--------------------------------------"
	@echo ">   OK: Postgres stopped and data destroyed."
	@echo "--------------------------------------"

## Open a psql session against the local dev database.
db/psql:
	@docker compose exec postgres psql -U starter -d starter

# ---------------------------------------------------------------------------
# Setup (install dev tools)
# ---------------------------------------------------------------------------

## Install development tools: golangci-lint, air.
setup:
	@echo "--------------------------------------"
	@echo ">   Installing golangci-lint..."
	@echo "--------------------------------------"
	@if command -v brew > /dev/null 2>&1; then \
		brew install golangci-lint; \
	else \
		echo ">   Homebrew not found. Install golangci-lint manually:"; \
		echo ">   https://golangci-lint.run/welcome/install/"; \
	fi
	@echo "--------------------------------------"
	@echo ">   Installing air (live-reload)..."
	@echo "--------------------------------------"
	@go install github.com/air-verse/air@latest
	@echo "--------------------------------------"
	@echo ">   OK: Setup complete."
	@echo ">   Make sure $$(go env GOPATH)/bin is in your PATH for air."
	@echo "--------------------------------------"

# ---------------------------------------------------------------------------
# Docker image
# ---------------------------------------------------------------------------

image: build
	@command -v docker > /dev/null 2>&1 || (echo "Docker not found. Please install Docker." && exit 1)
	@echo "Building Docker image: $(IMAGE_NAME):$(TAG)"
	@docker build -t $(IMAGE_NAME):$(TAG) -t $(IMAGE_NAME):latest .
ifeq ($(HAS_KIND),yes)
	@echo "--------------------------------------"
	@echo ">   Pushing image to 'kind' registry..."
	@echo "--------------------------------------"
	@kind load docker-image $(IMAGE_NAME):$(TAG)
endif
	@echo "--------------------------------------"
	@echo ">   OK: Image $(IMAGE_NAME):$(TAG) built."
	@echo "--------------------------------------"

# ---------------------------------------------------------------------------
# Helm / Kubernetes
# ---------------------------------------------------------------------------

chart: image
	@if [ -f helm/values.yaml ]; then \
		sed -i.bak "s|repository:.*|repository: $(IMAGE_NAME)|" helm/values.yaml; \
		sed -i.bak "s|tag:.*|tag: \"latest\"|"             helm/values.yaml; \
		sed -i.bak "s|pullPolicy:.*|pullPolicy: IfNotPresent|" helm/values.yaml; \
		rm -f helm/values.yaml.bak; \
		echo "--------------------------------------"; \
		echo ">   OK: helm/values.yaml updated."; \
		echo "--------------------------------------"; \
	else \
		echo ">   WARNING: helm/values.yaml not found."; \
	fi

up: chart
	@echo "--------------------------------------"
	@echo ">   Deploying to Kubernetes via Helm..."
	@echo "--------------------------------------"
	@if [ -d helm ]; then \
		kubectl create namespace $(NAMESPACE) 2>/dev/null || true; \
		helm upgrade --install $(APP_NAME) helm \
			--namespace $(NAMESPACE) \
			--set image.repository=$(IMAGE_NAME) \
			--set image.tag=$(TAG); \
		echo "--------------------------------------"; \
		echo ">   OK: Deployment complete."; \
		echo "--------------------------------------"; \
	else \
		echo ">   ERROR: helm directory not found."; \
		exit 1; \
	fi

down:
	@echo "--------------------------------------"
	@echo ">   Uninstalling Helm release..."
	@echo "--------------------------------------"
	@helm uninstall $(APP_NAME) --namespace $(NAMESPACE) || true
	@echo "--------------------------------------"
	@echo ">   OK: Uninstallation complete."
	@echo "--------------------------------------"

# ---------------------------------------------------------------------------
# Help
# ---------------------------------------------------------------------------

help:
	@echo ""
	@echo "go-service-starter — available targets"
	@echo ""
	@echo "  Development"
	@echo "    make build        Build the binary to out/main"
	@echo "    make run          Build and run the server (reads .env)"
	@echo "    make watch        Live-reload on source changes (requires make setup)"
	@echo "    make test         Run all tests with race detector"
	@echo "    make coverage     Run tests and produce coverage summary + HTML report"
	@echo "    make fmt          Auto-format all Go source files"
	@echo "    make lint         go vet + golangci-lint (if installed)"
	@echo "    make clean        Remove build artifacts and Docker image"
	@echo "    make setup        Install golangci-lint and air via brew/go install"
	@echo ""
	@echo "  Local database"
	@echo "    make db/up        Start Postgres via docker compose"
	@echo "    make db/down      Stop Postgres (data volume preserved)"
	@echo "    make db/destroy   Stop Postgres and delete all data (removes volume)"
	@echo "    make db/psql      Open psql session against local dev DB"
	@echo ""
	@echo "  Docker / Kubernetes"
	@echo "    make image        Build Docker image (+ load into kind if available)"
	@echo "    make chart        Update helm/values.yaml and build image"
	@echo "    make up           Deploy to Kubernetes via Helm"
	@echo "    make down         Remove Helm release"
	@echo ""
	@echo "  Quick start:"
	@echo "    cp .env.example .env   # fill in DATABASE_URL"
	@echo "    make db/up             # start local Postgres"
	@echo "    make run               # build and run (migrations run automatically)"
	@echo ""
