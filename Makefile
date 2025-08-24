# Makefile for gh-wizard

.PHONY: install setup dev build test clean help

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

install: ## Install gh-wizard as GitHub CLI extension
	gh extension install .

setup: install setup-hooks ## Complete setup: install extension + git hooks
	@echo "✅ Setup complete!"
	@echo "   - gh-wizard installed as GitHub CLI extension"
	@echo "   - Git hooks configured (if lefthook available)"
	@echo ""
	@echo "Usage:"
	@echo "  gh wizard                    # Run wizard"
	@echo "  gh wizard --help            # Show help"

setup-hooks: ## Setup git hooks (requires lefthook)
	@if command -v lefthook >/dev/null 2>&1; then \
		echo "🪝 Installing git hooks with lefthook..."; \
		lefthook install; \
		echo "✅ Git hooks installed"; \
	else \
		echo "⚠️  lefthook not found. Install it for automatic conventional commits:"; \
		echo "   macOS: brew install lefthook"; \
		echo "   Linux: go install github.com/evilmartians/lefthook@latest"; \
		echo "   Windows: scoop install lefthook"; \
		echo "   Then run: make setup-hooks"; \
	fi

dev: ## Development setup: install + hooks + tests
	@$(MAKE) setup
	@$(MAKE) test

build: ## Build the binary
	go build -o gh-wizard

test: ## Run tests
	go test ./...

test-verbose: ## Run tests with verbose output
	go test -v ./...

fmt: ## Format code
	go fmt ./...

lint: ## Run linter (requires golangci-lint)
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not found. Install it for linting:"; \
		echo "   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

clean: ## Clean build artifacts
	rm -f gh-wizard
	go clean

# Development workflow targets
commit-check: fmt test ## Pre-commit checks
	@echo "✅ Code formatted and tests passed"

release-check: commit-check lint ## Pre-release checks
	@echo "✅ Ready for release"

# Docker support (optional)
docker-build: ## Build Docker image
	docker build -t gh-wizard .

docker-run: ## Run in Docker container
	docker run --rm -it gh-wizard

# Help lefthook users
hooks-status: ## Check git hooks status
	@if command -v lefthook >/dev/null 2>&1; then \
		echo "🪝 Lefthook status:"; \
		lefthook version; \
		echo ""; \
		echo "Configured hooks:"; \
		ls -la .git/hooks/ | grep -v sample || echo "No hooks found"; \
	else \
		echo "⚠️  lefthook not installed"; \
	fi

uninstall-hooks: ## Remove git hooks
	@if command -v lefthook >/dev/null 2>&1; then \
		lefthook uninstall; \
		echo "✅ Git hooks removed"; \
	else \
		echo "ℹ️  lefthook not found, removing hooks manually..."; \
		rm -f .git/hooks/commit-msg .git/hooks/pre-commit; \
		echo "✅ Git hooks removed"; \
	fi