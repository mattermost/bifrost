.PHONY: build check-style install run test verify-gomod

## Docker Build Versions
DOCKER_BUILD_IMAGE = golang:1.17.3
DOCKER_BASE_IMAGE = alpine:3.14
# Build variables
COMMIT_HASH  ?= $(shell git rev-parse HEAD)
BUILD_DATE   ?= $(shell date +%FT%T%z)

# Release variables
TAG_EXISTS=$(shell git rev-parse $(NEXT_VER) >/dev/null 2>&1; echo $$?)

# Variables
GO=go
APP:=bifrost
APPNAME:=bifrost
MATTERMOST_BIFROST_IMAGE ?= mattermost/bifrost:test

# Flags
LDFLAGS :="
LDFLAGS += -X github.com/mattermost/$(APP)/internal/server.CommitHash=$(COMMIT_HASH)
LDFLAGS += -X github.com/mattermost/$(APP)/internal/server.BuildDate=$(BUILD_DATE)
LDFLAGS +="

# Build for distribution
build:
	@echo Building Mattermost Bifrost
	env GOOS=linux GOARCH=amd64 $(GO) build -ldflags $(LDFLAGS) -o $(APPNAME) ./cmd/$(APP)

# Builds the docker image
build-image:
	@echo Building Mattermost Bifrost Docker Image
	docker build \
	--build-arg DOCKER_BUILD_IMAGE=$(DOCKER_BUILD_IMAGE) \
	--build-arg DOCKER_BASE_IMAGE=$(DOCKER_BASE_IMAGE) \
	. -f build/Dockerfile \
	-t $(MATTERMOST_BIFROST_IMAGE) \
	--no-cache

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
