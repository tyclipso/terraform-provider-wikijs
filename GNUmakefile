SHELL := /usr/bin/env bash

BINARY  := terraform-provider-wikijs
BIN_DIR := $(CURDIR)/bin

GOOS     := $(shell go env GOOS)
GOARCH   := $(shell go env GOARCH)
PLATFORM := $(GOOS)_$(GOARCH)

# Version stamped into the dev build consumed via dev_overrides.
DEV_VERSION ?= 0.0.0-dev

# Simply-expanded on purpose: `date` must run exactly once per make
# invocation, otherwise targets in the same run disagree on the path.
TIMESTAMP      := $(shell date +%s)
MIRROR_VERSION ?= 99.0.$(TIMESTAMP)
MIRROR_ROOT    ?= $(HOME)/.terraform.d/plugins/registry.terraform.io/tyclipso/wikijs
MIRROR_DIR      = $(MIRROR_ROOT)/$(MIRROR_VERSION)/$(PLATFORM)
MIRROR_KEEP    ?= 7

# Set variables for terraform.local registry
# PROVIDER_VERSION needs to be set outside
TERRAFORM_LOCAL_VERSION  = $(PROVIDER_VERSION)
TERRAFORM_LOCAL_ROOT    ?= $(HOME)/.terraform.d/plugins/terraform.local/tyclipso/wikijs
TERRAFORM_LOCAL_DIR      = $(TERRAFORM_LOCAL_ROOT)/$(TERRAFORM_LOCAL_VERSION)/$(PLATFORM)

# Set directory for tool binaries and make it available in PATH
TOOL_DIR := $(CURDIR)/bin/tools
export PATH := $(TOOL_DIR):$(PATH)

# Keep high dependency tools out of the package chain and instead install them
# directly to a pinned version for easier make handling
GOLANGCI_LINT_VERSION ?= v2.13.2
GORELEASER_VERSION    ?= v2.18.0

GOLANGCI_LINT_MOD := github.com/golangci/golangci-lint/v2
GORELEASER_MOD    := github.com/goreleaser/goreleaser/v2

GOLANGCI_LINT_PKG := $(GOLANGCI_LINT_MOD)/cmd/golangci-lint
GORELEASER_PKG    := $(GORELEASER_MOD)

# Ensure presence of tool bin dir
$(TOOL_DIR):
	mkdir -p $@

# Stamp files: bumping a version variable invalidates the stamp and
# triggers a reinstall, which a plain binary target would not.
$(TOOL_DIR)/.golangci-lint-$(GOLANGCI_LINT_VERSION): | $(TOOL_DIR)
	GOBIN=$(TOOL_DIR) go install $(GOLANGCI_LINT_PKG)@$(GOLANGCI_LINT_VERSION)
	rm -f $(TOOL_DIR)/.golangci-lint-*
	touch $@

$(TOOL_DIR)/.goreleaser-$(GORELEASER_VERSION): | $(TOOL_DIR)
	GOBIN=$(TOOL_DIR) go install $(GORELEASER_PKG)@$(GORELEASER_VERSION)
	rm -f $(TOOL_DIR)/.goreleaser-*
	touch $@

# Package holding the //go:generate directive for tfplugindocs.
DOCS_PKG ?= .

TESTARGS ?=

.DEFAULT_GOAL := build

## help: list available targets
.PHONY: help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## /  /'

## build: compile the dev binary into ./bin for dev_overrides
.PHONY: build
build:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build \
		-ldflags="-X 'main.Version=$(DEV_VERSION)'" \
		-o $(BIN_DIR)/$(BINARY) .

## build-debug: compile with optimisations off, for delve
.PHONY: build-debug
build-debug:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build \
		-gcflags="all=-N -l" \
		-ldflags="-X 'main.Version=$(DEV_VERSION)'" \
		-o $(BIN_DIR)/$(BINARY) .

## debug: run the provider in debug mode; prints TF_REATTACH_PROVIDERS
.PHONY: debug
debug: build-debug
	$(BIN_DIR)/$(BINARY) -debug

## dlv: run the provider under delve in debug mode
.PHONY: dlv
dlv: build-debug $(TOOL_DIR)/dlv
	dlv exec $(BIN_DIR)/$(BINARY) -- -debug

## tool: list developer tools, pinned versions and install state
.PHONY: tool
tool:
	@row() { \
		name="$$1"; pinned="$$2"; \
		if [ -x "$(TOOL_DIR)/$$name" ]; then \
			installed=$$(go version -m "$(TOOL_DIR)/$$name" 2>/dev/null | awk '$$1=="mod"{print $$3; exit}'); \
			if [ "$$installed" = "$$pinned" ]; then state="built"; \
			else state="stale: $$installed"; fi; \
		else \
			state="not built"; \
		fi; \
		printf '  %-16s %-12s %s\n' "$$name" "$$pinned" "$$state"; \
	}; \
	echo "Module tools (go.mod tool block):"; \
	go list -f '{{with .Module}}{{.Version}}{{end}} {{.ImportPath}}' tool | \
	while read -r version pkg; do \
		name="$${pkg##*/}"; \
		case "$$name" in v[0-9]*) rest="$${pkg%/*}"; name="$${rest##*/}";; esac; \
		row "$$name" "$$version"; \
	done; \
	echo; \
	echo "Standalone tools (pinned in $(firstword $(MAKEFILE_LIST))):"; \
	row golangci-lint "$(GOLANGCI_LINT_VERSION)"; \
	row goreleaser "$(GORELEASER_VERSION)"

## tools: build module-pinned tools and install standalone ones
.PHONY: tools
tools: tools-mod $(TOOL_DIR)/.golangci-lint-$(GOLANGCI_LINT_VERSION) $(TOOL_DIR)/.goreleaser-$(GORELEASER_VERSION)

.PHONY: tools-mod
tools-mod: | $(TOOL_DIR)
	go build -o $(TOOL_DIR)/ tool

# $(1) display name, $(2) module path, $(3) currently pinned version
define check_tool
	latest=$$(go list -m -f '{{.Version}}' $(2)@latest 2>/dev/null); \
	if [ -z "$$latest" ]; then \
		echo "  $(1): could not query $(2)"; \
	elif [ "$$latest" = "$(3)" ]; then \
		echo "  $(1): $(3) (current)"; \
	else \
		echo "  $(1): $(3) -> $$latest"; \
	fi
endef

## tools-outdated: report newer releases of the standalone pinned tools
.PHONY: tools-outdated
tools-outdated:
	@echo "Standalone tools pinned in $(firstword $(MAKEFILE_LIST)):"
	@$(call check_tool,golangci-lint,$(GOLANGCI_LINT_MOD),$(GOLANGCI_LINT_VERSION))
	@$(call check_tool,goreleaser,$(GORELEASER_MOD),$(GORELEASER_VERSION))


## tools-update: bump all tool dependencies
.PHONY: tools-update
tools-update:
	go get tool
	go mod tidy
	@echo
	@$(MAKE) --no-print-directory tools-outdated

## generate: run all code and doc generators
.PHONY: generate
generate: graphql docs

## graphql: regenerate the genqlient bindings in ./wikijs
.PHONY: graphql
graphql: $(TOOL_DIR)/genqlient
	go generate ./wikijs

## docs: regenerate the registry documentation
.PHONY: docs
docs: $(TOOL_DIR)/tfplugindocs
	go generate $(DOCS_PKG)

## check-generate: fail if generated files are not committed
.PHONY: check-generate
check-generate: generate
	git diff --exit-code -- docs wikijs examples

## terraform.local: build a versioned package into the terraform.local filesystem_mirror
.PHONY: terraform.local
terraform.local:
	mkdir -p $(TERRAFORM_LOCAL_DIR)
	CGO_ENABLED=0 go build \
		-ldflags="-X 'main.Version=$(TERRAFORM_LOCAL_VERSION)'" \
		-o $(TERRAFORM_LOCAL_DIR)/$(BINARY)_v$(TERRAFORM_LOCAL_VERSION) .
	@echo
	@echo "Installed $(TERRAFORM_LOCAL_VERSION) for $(PLATFORM)."
	@echo "In the testbed: rewrite all `required_providers`, to use the new registry terraform.local"
	@echo "After that to transfer existing ressources to the new provider, run"
	@echo "`terraform state replace-provider tyclipso/wikijs terraform.local/tyclipso/wikijs`"
	@echo "Skip if starting from clean slate. Finally run `terraform init`"

## terraform.local-clean: remove all terraform.local builds
.PHONY: terraform.local-clean
terraform.local-clean:
	@test -d $(TERRAFORM_LOCAL_ROOT) || { echo "no mirror at $(TERRAFORM_LOCAL_ROOT)"; exit 0; }
	find $(TERRAFORM_LOCAL_ROOT) -mindepth 1 -maxdepth 1 -type d \
		-name '*.*.*' -exec rm -rf {} +

## mirror: build a versioned package into the local filesystem_mirror
.PHONY: mirror
mirror:
	mkdir -p $(MIRROR_DIR)
	CGO_ENABLED=0 go build \
		-ldflags="-X 'main.Version=$(MIRROR_VERSION)'" \
		-o $(MIRROR_DIR)/$(BINARY)_v$(MIRROR_VERSION) .
	@echo
	@echo "Installed $(MIRROR_VERSION) for $(PLATFORM)."
	@echo "In the testbed: comment out `dev_overrides`, then `terraform init -upgrade`"

## mirror-clean: drop mirror builds older than MIRROR_KEEP days
.PHONY: mirror-clean
mirror-clean:
	@test -d $(MIRROR_ROOT) || { echo "no mirror at $(MIRROR_ROOT)"; exit 0; }
	find $(MIRROR_ROOT) -mindepth 1 -maxdepth 1 -type d \
		-name '99.0.*' -mtime +$(MIRROR_KEEP) -exec rm -rf {} +

## clean: remove build artifacts and prune old mirror builds
.PHONY: clean
clean: mirror-clean terraform.local-clean
	rm -f $(BIN_DIR)/$(BINARY)

## clean-tools: remove built developer tools
.PHONY: clean-tools
clean-tools:
	rm -rf $(TOOL_DIR)

## binclean: clean plus developer tools
.PHONY: binclean
distclean: clean clean-tools
	rm -rf $(BIN_DIR)

## fmt: gofmt the tree
.PHONY: fmt
fmt:
	gofmt -s -w -l .

## vet: run go vet
.PHONY: vet
vet:
	go vet ./...

## lint: run golangci-lint using the repo config
.PHONY: lint
lint: $(TOOL_DIR)/.golangci-lint-$(GOLANGCI_LINT_VERSION)
	golangci-lint run

## tidy: tidy and verify module dependencies
.PHONY: tidy
tidy:
	go mod tidy
	go mod verify
	
## update: bump dependencies to their latest minor/patch releases
.PHONY: update
update:
	go get -u ./...
	go mod tidy

## update-patch: bump dependencies to their latest patch releases only
.PHONY: update-patch
update-patch:
	go get -u=patch ./...
	go mod tidy

## test: unit tests only, no acceptance tests
.PHONY: test
test:
	go test ./... $(TESTARGS) -timeout 5m

## testacc: full acceptance suite against a live Wiki.js
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

## snapshot: build a release package locally without tagging
.PHONY: snapshot
snapshot: $(TOOL_DIR)/.goreleaser-$(GORELEASER_VERSION)
	goreleaser release --snapshot --clean

## check: everything CI should agree with, minus acceptance tests
.PHONY: check
check: fmt vet lint test check-generate
