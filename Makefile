# Define the golangci-lint version
GOLANGCI_LINT_VERSION = v1.61.0

# Path to the golangci-lint binary
GOLANGCI_LINT_BIN = $(shell go env GOPATH)/bin/golangci-lint

# Path to the air binary
AIR_BIN = $(shell go env GOPATH)/bin/air

SWAG_BIN = $(shell go env GOPATH)/bin/swag

# Ensure golangci-lint is installed
.PHONY: ensure-golangci-lint
ensure-golangci-lint:
	@which $(GOLANGCI_LINT_BIN) > /dev/null || { \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin $(GOLANGCI_LINT_VERSION); \
	}

# Ensure the air binary is installed
.PHONY: ensure-air
ensure-air:
	@which air > /dev/null || { \
		echo "installing air..."; \
		go install github.com/air-verse/air@latest; \
	}

.PHONY: ensure-swag
ensure-swag:
	@which swag > /dev/null || { \
		echo "installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	}

run: ensure-air
	$(AIR_BIN)

lint: ensure-golangci-lint
	$(GOLANGCI_LINT_BIN) run ./...

test:
	go test -v ./...

doc: ensure-swag
	$(SWAG_BIN) init -g ./internal/api/router/router.go