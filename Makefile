# Go build configuration
BINARY_NAME=mmd-cli
GO=/usr/local/go/bin/go
VERSION?=$(shell cat version)
LDFLAGS=-ldflags '-s -w -X github.com/coolamit/mermaid-cli/internal/cli.Version=$(VERSION)'

# Declare phony targets
.PHONY: ssh-cmd up down build clean test tidy build-linux-x64 build-linux-arm64 build-macos-x64 build-macos-arm64 build-all docker-up docker-down docker-run docker-run-aloof docker-clean

# Common function definitions
define CURRENT_HOMESTEAD_STATUS
cd ../.. && \
if vagrant status | grep -qE "running \(virtualbox\)|running \(parallels\)"; then echo "running"; else echo "stopped"; fi
endef

define SSH_EXEC
PROJECT_NAME=$$(basename $$(pwd)) && \
(cd ../.. && vagrant ssh -- -t "cd code/$$PROJECT_NAME && $(1)")
endef

# This command will start Homestead if it is not already running.
up:
	@echo "Checking Homestead status..."
	@STATUS=$$($(CURRENT_HOMESTEAD_STATUS)) && \
	if [ "$$STATUS" = "running" ]; then \
		echo "Homestead is already running."; \
	else \
		echo "Starting Homestead..." && cd ../.. && vagrant up; \
	fi

# This command will stop Homestead if it is running.
down:
	@echo "Checking Homestead status..."
	@STATUS=$$($(CURRENT_HOMESTEAD_STATUS)) && \
	if [ "$$STATUS" = "running" ]; then \
		echo "Stopping Homestead..." && cd ../.. && vagrant halt; \
	else \
		echo "Homestead is NOT currently running."; \
	fi

# Command to allow any command to be run in Homestead VM
# Usage: use -- separator before the command
# Example: make ssh-cmd -- ls -alt
#        : make ssh-cmd -- mod tidy
ssh-cmd:
	@$(call SSH_EXEC,$(filter-out $@,$(MAKECMDGOALS)))

# The wildcard rule is needed to allow for artisan or other ssh commands which can have any name.
# This ensures that the catch-all only matches targets that come after
# 'ssh-cmd' in the command line.
ifneq (,$(filter $(firstword $(MAKECMDGOALS)),ssh-cmd docker-run docker-run-aloof))
%:
	@:
endif

# Go build targets (run inside Homestead VM)
build:
	@$(call SSH_EXEC,$(GO) build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/mmd-cli)
	@echo "Built $(BINARY_NAME)"

test:
	@$(call SSH_EXEC,$(GO) clean -testcache && $(GO) test ./...)

tidy:
	@$(call SSH_EXEC,$(GO) mod tidy)

clean:
	@$(call SSH_EXEC,rm -f $(BINARY_NAME))
	@echo "Cleaned"

# Cross-compilation targets
build-linux-x64:
	@$(call SSH_EXEC,GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BINARY_NAME)-linux-x64 ./cmd/mmd-cli)

build-linux-arm64:
	@$(call SSH_EXEC,GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BINARY_NAME)-linux-arm64 ./cmd/mmd-cli)

build-macos-x64:
	@$(call SSH_EXEC,GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BINARY_NAME)-macos-x64 ./cmd/mmd-cli)

build-macos-arm64:
	@$(call SSH_EXEC,GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BINARY_NAME)-macos-arm64 ./cmd/mmd-cli)

build-all: build-linux-x64 build-linux-arm64 build-macos-x64 build-macos-arm64
	@echo "Built all platforms"

# Docker Compose targets (run on host, not in Homestead VM)
docker-up:
	docker compose up -d

docker-down:
	docker compose stop

# Run a command in the running Docker container.
# Example: make docker-run -- mmd-cli -i diagram.mmd -o diagram.svg
docker-run:
	docker compose exec mmd-cli $(filter-out $@,$(MAKECMDGOALS))

# Run mmd-cli in an ephemeral container (no docker-up needed).
# Example: make docker-run-aloof -- -i diagram.mmd -o diagram.svg
docker-run-aloof:
	docker compose run --rm --entrypoint mmd-cli mmd-cli $(filter-out $@,$(MAKECMDGOALS))

docker-clean:
	docker compose down --rmi local --volumes --remove-orphans
