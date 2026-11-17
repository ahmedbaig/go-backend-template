APP        ?= $(notdir $(CURDIR))
CMD        ?= ./cmd/server
VERSION    ?= $(shell cat VERSION 2>/dev/null || echo 0.0.0)
CONFIG_DIR ?= $(CURDIR)/var
IMAGE      ?= $(APP)
PORT       ?= 8080

.PHONY: help fmt vet lint test cover build run docker docker-run tidy vuln check clean

help:
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

fmt: ## gofmt everything
	gofmt -w .

vet: ## go vet
	go vet ./...

lint: vet ## vet + staticcheck
	go run honnef.co/go/tools/cmd/staticcheck@latest ./...

test: ## tests with race detector and an empty CONFIG_DIR
	CONFIG_DIR=$$(mktemp -d) go test -race ./...

cover: ## tests with coverage summary
	CONFIG_DIR=$$(mktemp -d) go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out | tail -1

vuln: ## govulncheck
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

check: lint test vuln ## what CI runs

tidy: ## go mod tidy
	go mod tidy

build: ## build bin/$(APP)
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/$(APP) $(CMD)

run: ## run locally with CONFIG_DIR=./var
	CONFIG_DIR=$(CONFIG_DIR) go run $(CMD)

docker: ## build $(IMAGE):local
	docker build --build-arg CMD=$(CMD) -t $(IMAGE):local .

docker-run: docker ## run $(IMAGE):local with ./var as config
	docker run --rm -p $(PORT):$(PORT) -e PORT=$(PORT) -v $(CONFIG_DIR):/app/var/secrets:ro $(IMAGE):local

clean: ## delete bin/ and coverage.out
	go clean
	$(RM) -r bin
	$(RM) coverage.out
