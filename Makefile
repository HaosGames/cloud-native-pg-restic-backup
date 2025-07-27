.PHONY: build test clean docker-build docker-push

# Variables
PLUGIN_IMAGE ?= cnpg-restic-plugin
TAG ?= latest
GO ?= go

# Build the plugin binary
build:
	$(GO) build -o bin/plugin ./cmd/plugin

# Run tests
test:
	$(GO) test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Build Docker image
docker-build:
	docker build -t $(PLUGIN_IMAGE):$(TAG) .

# Push Docker image
docker-push:
	docker push $(PLUGIN_IMAGE):$(TAG)

# Run local development setup
setup-dev:
	chmod +x setup-dev.sh
	./setup-dev.sh

# Deploy test configuration to kind cluster
deploy-test:
	kubectl apply -f examples/plugin-config.yaml

# Get plugin logs
logs:
	kubectl logs -n cnpg-system -l app=postgresql-test -c backup --tail=100 -f

# Create a temporary test container with restic
test-restic:
	docker run --rm -it \
		-e RESTIC_REPOSITORY \
		-e RESTIC_PASSWORD \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e AWS_ENDPOINT \
		ubuntu:22.04 \
		bash -c "apt-get update && apt-get install -y restic && bash"
