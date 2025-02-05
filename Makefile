# Make settings
# Mostly copied from: https://tech.davis-hansson.com/p/make/

# Use Bash
SHELL := bash

# If one of the commands fails just fail properly and don't run the other commands.
.SHELLFLAGS := -eu -o pipefail -c

# Allows me to use a single shell session so you can do things like 'cd' without doing hacks.
.ONESHELL:

# Tells make not to do crazy shit.
MAKEFLAGS += --no-builtin-rules

# Allows me to replace tabs with > characters. This makes the things a bit easier to use things like forloops in bash.
ifeq ($(origin .RECIPEPREFIX), undefined)
  $(error This Make does not support .RECIPEPREFIX. Please use GNU Make 4.0 or later)
endif
.RECIPEPREFIX = >


# App Vars

APP_NAME = rc3
SEMVER = 0.0.0

# Although go 1.18 has the git info baked into the binary now it still seems like there is no support
# For including outside variables except this. So keep it for now.
GO_LDFLAGS = '-X "github.com/clintjedwards/${APP_NAME}/internal/cli.appVersion=$(SEMVER)" \
				-X "github.com/clintjedwards/${APP_NAME}/internal/api.appVersion=$(SEMVER)"'
SHELL = /bin/bash

## build: run tests and compile application
build: check-path-included check-semver-included
> go test ./...
> go mod tidy
> export CGO_ENABLED=1
> go build -tags release -ldflags $(GO_LDFLAGS) -o $(OUTPUT)
.PHONY: build

## run: build application and run server with frontend
run:
> @$(MAKE) -j run-tailwind run-backend
.PHONY: build

## run-backend: build application and run server
run-backend:
> go build -ldflags $(GO_LDFLAGS) -o /tmp/${APP_NAME}
> /tmp/${APP_NAME} service start
.PHONY: run

## run-tailwind: watch and build tailwind assets
run-tailwind:
> npx tailwindcss@3.4.17 -i ./internal/frontend/main.css -o ./internal/frontend/public/css/main.css --watch &> /dev/null

## run-docs: build and run documentation website for development
run-docs:
> cd documentation
> mdbook serve --open
.PHONY: run-docs

## build-docs: build final documentation site artifacts
build-docs:
> cd documentation
> mdbook build
.PHONY: build-docs

## help: prints this help message
help:
> @echo "Usage: "
> @sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
.PHONY: help

check-path-included:
ifndef OUTPUT
>	$(error OUTPUT is undefined; ex. OUTPUT=/tmp/${APP_NAME})
endif

check-semver-included:
ifeq ($(SEMVER), 0.0.0)
>	$(error SEMVER is undefined; ex. SEMVER=0.0.1)
endif
