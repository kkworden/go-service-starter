.PHONY: all build clean up image chart test lint

IMAGE_NAME = go-service-starter
TAG = latest
NAMESPACE = default
APP_NAME = go-service

# This Makefile can specifically detect if you have "kind", a Kubernetes
# installation. If so, it will push the docker image to the kind registry,
# so you can deploy containers with your image.
HAS_KIND := $(shell command -v kind > /dev/null 2>&1 && echo "yes" || echo "no")


# Default target
all: build

test:
	go test -race ./...
	@echo "--------------------------------------";
	@echo ">   OK: Tests passed.";
	@echo "--------------------------------------";

lint:
	go vet ./...
	@echo "--------------------------------------";
	@echo ">   OK: Lint passed.";
	@echo "--------------------------------------";

build:
	mkdir -p out/
	go mod download
	go build -o out/main .
	@echo "--------------------------------------";
	@echo ">   OK: Build done.";
	@echo "--------------------------------------";

clean:
	rm -rf out/
	docker rmi -f $(IMAGE_NAME):$(TAG) || true

up: chart
	@echo "Deploying to Kubernetes..."
	@echo "--------------------------------------";
	@echo ">   Deploying to Kubernetes via Helm...";
	@echo "--------------------------------------";
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
		echo ">   ERROR: Helm chart directory not found."; \
		echo ">   Please create a helm directory with your chart."; \
		echo "--------------------------------------"; \
		exit 1; \
	fi

down:
	@echo "Removing Kubernetes deployment..."
	@echo "--------------------------------------";
	@echo ">   Uninstalling Helm release...";
	@echo "--------------------------------------";
	helm uninstall $(APP_NAME) --namespace $(NAMESPACE) || true
	@echo "--------------------------------------";
	@echo ">   OK: Uninstallation complete.";
	@echo "--------------------------------------";

image: build
	@command -v docker > /dev/null 2>&1 || (echo "Docker not found. Please install Docker." && exit 1)
	@echo "Building Docker image locally: $(IMAGE_NAME):$(TAG)"
	docker build -t $(IMAGE_NAME):$(TAG) -t $(IMAGE_NAME):latest .
ifeq ($(HAS_KIND),yes)
	@echo "--------------------------------------";
	@echo ">   Pushing image to 'kind' registry...";
	@echo "--------------------------------------";
	kind load docker-image $(IMAGE_NAME):$(TAG)
endif
	@echo "--------------------------------------";
	@echo ">   OK: Image $(IMAGE_NAME):$(TAG) built successfully."
	@echo "--------------------------------------";


chart: image
	@bash -c 'if [ -f helm/values.yaml ]; then \
		echo "--------------------------------------"; \
		echo ">   OK: helm/values.yaml was modified. Helm chart updated."; \
		echo "--------------------------------------"; \
		sed -i.bak "s|repository:.*|repository: go-service-starter|" helm/values.yaml; \
		sed -i.bak "s|tag:.*|tag: \"latest\"|" helm/values.yaml; \
		sed -i.bak "s|pullPolicy:.*|pullPolicy: IfNotPresent|" helm/values.yaml; \
		rm -f helm/values.yaml.bak; \
	else \
		echo "--------------------------------------"; \
		echo ">   WARNING: helm/values.yaml not found. No Helm chart was updated."; \
		echo "--------------------------------------"; \
	fi'