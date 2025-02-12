PROJECT_FULL_NAME := controller-utils
REPO_ROOT := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
EFFECTIVE_VERSION := $(shell $(REPO_ROOT)/hack/get-version.sh)

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

ROOT_CODE_DIRS := $(REPO_ROOT)/pkg/...

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: tidy
tidy: ## Runs 'go mod tidy' for all modules in this repo.
	@$(REPO_ROOT)/hack/tidy.sh

.PHONY: format
format: goimports ## Formats the imports.
	@FORMATTER=$(FORMATTER) $(REPO_ROOT)/hack/format.sh $(ROOT_CODE_DIRS)

.PHONY: verify
verify: golangci-lint goimports ## Runs linter, 'go vet', and checks if the formatter has been run.
	@( echo "> Verifying root module ..." && \
		pushd $(REPO_ROOT) &>/dev/null && \
		go vet $(ROOT_CODE_DIRS) && \
		$(LINTER) run -c $(REPO_ROOT)/.golangci.yaml $(ROOT_CODE_DIRS) && \
		popd &>/dev/null )
	@test "$(SKIP_FORMATTING_CHECK)" = "true" || \
		( echo "> Checking for unformatted files ..." && \
		FORMATTER=$(FORMATTER) $(REPO_ROOT)/hack/format.sh --verify $(ROOT_CODE_DIRS) )

.PHONY: test
test: ## Run tests.
	go test $(ROOT_CODE_DIRS) -coverprofile cover.out
	go tool cover --html=cover.out -o cover.html
	go tool cover -func cover.out | tail -n 1

##@ Release

.PHONY: prepare-release
prepare-release: tidy format verify test

.PHONY: release-major
release-major: prepare-release ## Creates a new major release.
	@$(REPO_ROOT)/hack/release.sh major

.PHONY: release-minor
release-minor: prepare-release ## Creates a new minor release.
	@$(REPO_ROOT)/hack/release.sh minor

.PHONY: release-patch
release-patch: prepare-release ## Creates a new patch release.
	@$(REPO_ROOT)/hack/release.sh patch

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(REPO_ROOT)/bin

## Tool Binaries
FORMATTER ?= $(LOCALBIN)/goimports
LINTER ?= $(LOCALBIN)/golangci-lint

## Tool Versions
FORMATTER_VERSION ?= v0.22.0
LINTER_VERSION ?= 1.61.0

.PHONY: localbin
localbin:
	@test -d $(LOCALBIN) || mkdir -p $(LOCALBIN)

.PHONY: goimports
goimports: localbin ## Download goimports locally if necessary. If wrong version is installed, it will be overwritten.
	@test -s $(FORMATTER) && test -s ./hack/goimports_version && cat ./hack/goimports_version | grep -q $(FORMATTER_VERSION) || \
	( echo "Installing goimports $(FORMATTER_VERSION) ..."; \
	GOBIN=$(LOCALBIN) go install golang.org/x/tools/cmd/goimports@$(FORMATTER_VERSION) && \
	echo $(FORMATTER_VERSION) > ./hack/goimports_version )

.PHONY: golangci-lint
golangci-lint: localbin ## Download golangci-lint locally if necessary. If wrong version is installed, it will be overwritten.
	@test -s $(LINTER) && $(LINTER) --version | grep -q $(LINTER_VERSION) || \
	( echo "Installing golangci-lint $(LINTER_VERSION) ..."; \
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(LOCALBIN) v$(LINTER_VERSION) )





## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
CONTROLLER_TOOLS_VERSION ?= v0.15.0

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)
