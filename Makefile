.PHONY: build check-style install run test verify-gomod

GOLANG_VERSION := $(shell cat go.mod | grep "^go " | cut -d " " -f 2)

## Docker Build Versions
DOCKER_BUILD_IMAGE = golang:$(GOLANG_VERSION)-alpine3.19
DOCKER_BASE_IMAGE  = gcr.io/distroless/static:nonroot

# Build variables
COMMIT_HASH  ?= $(shell git rev-parse HEAD)
BUILD_DATE   ?= $(shell date +%FT%T%z)
PACKAGES      = $(shell go list ./...)

# Release variables
TAG_EXISTS=$(shell git rev-parse $(NEXT_VER) >/dev/null 2>&1; echo $$?)

# Variables
GO                  = go
GOPATH              = $(shell $(GO) env GOPATH)
APP                := bifrost
APPNAME            := bifrost
BIFROST_IMAGE_NAME ?= mattermost/$(APPNAME)
BIFROST_IMAGE_TAG  ?= test
BIFROST_IMAGE      ?= $(BIFROST_IMAGE_NAME):$(BIFROST_IMAGE_TAG)

# Flags
LDFLAGS :="
LDFLAGS += -X github.com/mattermost/$(APP)/internal/server.CommitHash=$(COMMIT_HASH)
LDFLAGS += -X github.com/mattermost/$(APP)/internal/server.BuildDate=$(BUILD_DATE)
LDFLAGS +="

TEST_FLAGS ?= -v

# Tools
GOLANGCILINT_VER := v1.56.2
GOLANGCILINT := $(TOOLS_BIN_DIR)/$(GOLANGCILINT_BIN)

TRIVY_SEVERITY := CRITICAL
TRIVY_EXIT_CODE := 1
TRIVY_VULN_TYPE := os,library

# Build for distribution
build:
	@echo Building Mattermost Bifrost
	env GOOS=linux GOARCH=amd64 $(GO) build -ldflags $(LDFLAGS) -o $(APPNAME) ./cmd/$(APP)

.PHONE: buildx-image
buildx-image:  ## Builds and pushes the docker image
	DOCKERFILE_PATH=build/Dockerfile BUILD_IMAGE=$(DOCKER_BUILD_IMAGE) BASE_IMAGE=$(DOCKER_BASE_IMAGE) IMAGE_NAME=$(BIFROST_IMAGE) ./scripts/build_image.sh buildx

.PHONE: build-image
build-image:  ## Build the docker image
	DOCKERFILE_PATH=build/Dockerfile BUILD_IMAGE=$(DOCKER_BUILD_IMAGE) BASE_IMAGE=$(DOCKER_BASE_IMAGE) IMAGE_NAME=$(BIFROST_IMAGE) ./scripts/build_image.sh local

# Build and install for the current platform
install:
	$(GO) install -ldflags $(LDFLAGS) ./cmd/$(APP)

# Run starts the app
run:
	$(GO) run ./cmd/$(APP)

# Test runs go test command
.PHONY: unittest
unittest:
	$(GO) test -cover -race -covermode=atomic -coverprofile=coverage.out $(TEST_FLAGS) ./...

## Runs govet and gofmt against all packages.
.PHONY: check-style
check-style: lint
	@echo Checking for style guide compliance

## Runs lint against all packages.
lint: $(GOPATH)/bin/golangci-lint
	@echo Running golangci-lint
	golangci-lint run

## Runs lint against all packages for changes only
lint-changes: $(GOPATH)/bin/golangci-lint
	@echo Running golangci-lint over changes only
	golangci-lint run -n

# Check modules
verify-gomod:
	$(GO) mod download
	$(GO) mod verify

# Draft a release
release:
	@if [[ -z "${NEXT_VER}" ]]; then \
		echo "Error: NEXT_VER must be defined, e.g. \"make release NEXT_VER=v1.0.1\""; \
		exit -1; \
	else \
		if [[ "${TAG_EXISTS}" -eq 0 ]]; then \
		  echo "Error: tag ${NEXT_VER} already exists"; \
			exit -1; \
		else \
			if ! [ -x "$$(command -v goreleaser)" ]; then \
			echo "goreleaser is not installed, do you want to download it? [y/N] " && read ans && [ $${ans:-N} = y ]; \
				if [ $$ans = y ] || [ $$ans = Y ]  ; then \
					curl -sfL https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh | sh; \
				else \
					echo "aborting make release."; \
					exit -1; \
				fi; \
			fi; \
			git commit -a -m 'Releasing $(NEXT_VER)'; \
			git tag $(NEXT_VER); \
			goreleaser --rm-dist; \
		fi; \
	fi;\

trivy: build-image ## Checks for vulnerabilities in the docker image
	@echo running trivy
	@trivy image --format table --exit-code $(TRIVY_EXIT_CODE) --ignore-unfixed --vuln-type $(TRIVY_VULN_TYPE) --severity $(TRIVY_SEVERITY) $(BIFROST_IMAGE)

.PHONY: clean
clean: ## Cleans the project and removes all generated files
	@rm -rf $(TOOLS_BIN_DIR)

## --------------------------------------
## Tooling Binaries
## --------------------------------------

$(GOPATH)/bin/golangci-lint: ## Install golangci-lint
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCILINT_VER)
