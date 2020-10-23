.PHONY: build check-style install run test verify-gomod

# Build variables
COMMIT_HASH  ?= $(shell git rev-parse HEAD)
BUILD_DATE   ?= $(shell date +%FT%T%z)

# Variables
GO=go
APP:=bifrost
APPNAME:=bifrost

# Flags
LDFLAGS :="
LDFLAGS += -X github.com/mattermost/$(APP)/internal/server.CommitHash=$(COMMIT_HASH)
LDFLAGS += -X github.com/mattermost/$(APP)/internal/server.BuildDate=$(BUILD_DATE)
LDFLAGS +="

# Build for distribution
build:
	@echo Default build Linux amd64
	env GOOS=linux GOARCH=amd64 $(GO) build -ldflags $(LDFLAGS) -o $(APPNAME) ./cmd/$(APP)

# Builds the docker image
docker: build
	docker build -t mattermost/$(APPNAME) .
	rm -f $(APPNAME)

# Build and install for the current platform
install:
	$(GO) install -ldflags $(LDFLAGS) ./cmd/$(APP)

# Run starts the app
run:
	$(GO) run ./cmd/$(APP)

# Test runs go test command
test: check-style
	$(GO) test -cover -race ./...

# Checks code style by running golangci-lint on codebase.
check-style:
	@if ! [ -x "$$(command -v golangci-lint)" ]; then \
		echo "golangci-lint is not installed. Please see https://github.com/golangci/golangci-lint#install for installation instructions."; \
		exit 1; \
	fi; \

	@echo Running golangci-lint
	golangci-lint run ./...

# Check modules
verify-gomod:
	$(GO) mod download
	$(GO) mod verify
