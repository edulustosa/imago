# Define the golangci-lint version
GOLANGCI_LINT_VERSION = v1.61.0

# Path to the golangci-lint binary
GOLANGCI_LINT_BIN = $(shell go env GOPATH)/bin/golangci-lint

# Path to the air binary
AIR_BIN = $(shell go env GOPATH)/bin/air

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

run: ensure-air
	$(AIR_BIN)

lint: ensure-golangci-lint
	$(GOLANGCI_LINT_BIN) run ./...

test:
	go test -v ./...