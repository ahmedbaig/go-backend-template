APP := go-backend-template
CMD := ./cmd/server

include make/go.mk

.PHONY: setup

setup: ## copy sample configs into ./var for make run
	mkdir -p var
	for f in configs/*.json.sample; do b=$$(basename $$f .sample); [ -f var/$$b ] || cp $$f var/$$b; done
	@echo "edit var/*.json, then: make run"
