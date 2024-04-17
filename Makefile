.PHONY: build check-style install run test verify-gomod

## Docker Build Versions
DOCKER_BUILD_IMAGE = golang:1.22-alpine3.19
DOCKER_BASE_IMAGE = alpine:3.19

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
BUILD_TIME := $(shell date -u +%Y%m%d.%H%M%S)
BUILD_HASH := $(shell git rev-parse HEAD)
ARCH ?= amd64

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
	@echo Building Mattermost Bifrost  for ARCH=$(ARCH)
	@if [ "$(ARCH)" = "amd64" ]; then \
		export GOARCH="amd64"; \
	elif [ "$(ARCH)" = "arm64" ]; then \
		export GOARCH="arm64"; \
	elif [ "$(ARCH)" = "arm" ]; then \
		export GOARCH="arm"; \
	else \
		echo "Unknown architecture $(ARCH)"; \
		exit 1; \
	fi; \
	env GOOS=linux $(GO) build -buildvcs=false -ldflags $(LDFLAGS) -gcflags all=-trimpath=$(PWD) -asmflags all=-trimpath=$(PWD) -a -installsuffix cgo -o build/_output/bin/$(APPNAME) ./cmd/$(APP)

.PHONY: build-image
build-image:  ## Build the docker image for Elrond
	@echo Building Elrond Docker Image
	@if [ -z "$(DOCKER_USERNAME)" ] || [ -z "$(DOCKER_PASSWORD)" ]; then \
		echo "DOCKER_USERNAME and/or DOCKER_PASSWORD not set. Skipping Docker login."; \
	else \
		echo $(DOCKER_PASSWORD) | docker login --username $(DOCKER_USERNAME) --password-stdin; \
	fi
	docker buildx build \
	--platform linux/arm64,linux/amd64 \
	--build-arg DOCKER_BUILD_IMAGE=$(DOCKER_BUILD_IMAGE) \
	--build-arg DOCKER_BASE_IMAGE=$(DOCKER_BASE_IMAGE) \
	. -f build/Dockerfile -t $(BIFROST_IMAGE) \
	--no-cache \
	--push

.PHONY: build-image-with-tag
build-image-with-tag:  ## Build the docker image for elrond
	@echo Building Elrond Docker Image
	@if [ -z "$(DOCKER_USERNAME)" ] || [ -z "$(DOCKER_PASSWORD)" ]; then \
		echo "DOCKER_USERNAME and/or DOCKER_PASSWORD not set. Skipping Docker login."; \
	else \
		echo $(DOCKER_PASSWORD) | docker login --username $(DOCKER_USERNAME) --password-stdin; \
	fi
	: $${TAG:?}
	docker buildx build \
	--platform linux/arm64,linux/amd64 \
	--build-arg DOCKER_BUILD_IMAGE=$(DOCKER_BUILD_IMAGE) \
	--build-arg DOCKER_BASE_IMAGE=$(DOCKER_BASE_IMAGE) \
	. -f build/Dockerfile -t $(BIFROST_IMAGE) -t $(BIFROST_IMAGE_NAME):${TAG} \
	--push

.PHONY: build-image-locally
build-image-locally:  ## Build the docker image for cloud-thanos-store-discovery
	@echo Building Biforst Docker Image
	@if [ -z "$(DOCKER_USERNAME)" ] || [ -z "$(DOCKER_PASSWORD)" ]; then \
		echo "DOCKER_USERNAME and/or DOCKER_PASSWORD not set. Skipping Docker login."; \
	else \
		echo $(DOCKER_PASSWORD) | docker login --username $(DOCKER_USERNAME) --password-stdin; \
	fi
	docker buildx build \
    --platform linux/arm64 \
	--build-arg DOCKER_BUILD_IMAGE=$(DOCKER_BUILD_IMAGE) \
	--build-arg DOCKER_BASE_IMAGE=$(DOCKER_BASE_IMAGE) \
	. -f build/Dockerfile -t $(BIFROST_IMAGE) \
	--no-cache \
	--load

.PHONY: scan
scan:
	docker scout cves $(BIFROST_IMAGE)

.PHONY: push-image-pr
push-image-pr:
	@echo Push Image PR
	./scripts/push-image-pr.sh

.PHONY: push-image
push-image:
	@echo Push Image
	./scripts/push-image.sh

.PHONY: install
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
check-style: govet lint
	@echo Checking for style guide compliance

## Runs lint against all packages.
lint: $(GOPATH)/bin/golangci-lint
	@echo Running golangci-lint
	golangci-lint run

## Runs lint against all packages for changes only
lint-changes: $(GOPATH)/bin/golangci-lint
	@echo Running golangci-lint over changes only
	golangci-lint run -n

## Runs govet against all packages.
.PHONY: govet
govet:
	@echo Running govet
	$(GO) vet ./...
	@echo Govet success

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
