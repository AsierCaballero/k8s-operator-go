SHELL := /bin/bash
.DEFAULT_GOAL := help

BIN_DIR := bin
BIN_NAME := manager
GO_MODULE := github.com/AsierCaballero/k8s-operator-go
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

TOOLS_DIR := $(BIN_DIR)
CONTROLLER_GEN := $(TOOLS_DIR)/controller-gen
SETUP_ENVTEST := $(TOOLS_DIR)/setup-envtest

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: all
all: test build ## Run tests and build binary

.PHONY: build
build: ## Build the manager binary
	go build -ldflags="-X main.Version=$(VERSION)" -o $(BIN_DIR)/$(BIN_NAME) cmd/manager/main.go

.PHONY: run
run: ## Run the operator locally
	go run cmd/manager/main.go

.PHONY: install
install: manifests ## Install CRDs into a cluster
	kubectl apply -f config/crd/

.PHONY: uninstall
uninstall: ## Uninstall CRDs from a cluster
	kubectl delete -f config/crd/ 2>/dev/null; true

.PHONY: deploy
deploy: build ## Deploy the operator to a cluster
	kubectl apply -k config/manager/

.PHONY: test
test: generate fmt vet ## Run tests
	go test ./... -v -coverprofile cover.out

.PHONY: lint
lint: ## Run linters
	golangci-lint run ./...

.PHONY: fmt
fmt: ## Format Go code
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: generate
generate: controller-gen ## Generate deepcopy and CRD manifests
	$(CONTROLLER_GEN) object:headerFile=hack/boilerplate.go.txt paths=./api/...
	$(CONTROLLER_GEN) crd:generateEmbeddedObjectMeta=true output:crd:dir=config/crd paths=./api/...
	$(CONTROLLER_GEN) webhook output:webhook:dir=config/webhook paths=./api/...

.PHONY: controller-gen
controller-gen: ## Install controller-gen
	GOBIN=$(abspath $(TOOLS_DIR)) go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.15.0

.PHONY: setup-envtest
setup-envtest: ## Install envtest binaries
	GOBIN=$(abspath $(TOOLS_DIR)) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: manifests
manifests: generate ## Generate all manifests

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf $(BIN_DIR) cover.out

.PHONY: docker-build
docker-build: ## Build the Docker image
	docker build -t ghcr.io/asiercaballero/k8s-operator-go:latest .

.PHONY: docker-push
docker-push: ## Push the Docker image
	docker push ghcr.io/asiercaballero/k8s-operator-go:latest

.PHONY: coverage
coverage: test ## Show test coverage
	go tool cover -html=cover.out -o cover.html
