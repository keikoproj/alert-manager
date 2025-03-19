# Image URL to use all building/pushing image targets
IMG ?=  keikoproj/alert-manager:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd"

# Tools required to run the full suite of tests properly
OSNAME           ?= $(shell uname -s | tr A-Z a-z)
KUBEBUILDER_VER  ?= 3.0.0
KUBEBUILDER_ARCH ?= amd64
ENVTEST_K8S_VERSION = 1.28.0
KIND_VERSION     ?= 0.20.0

# Tool Versions
CONTROLLER_TOOLS_VERSION ?= v0.13.0
KUSTOMIZE_VERSION ?= v3.8.7

# Kind configuration
KIND            ?= $(LOCALBIN)/kind
KIND_CLUSTER_NAME ?= alert-manager-test
KIND_K8S_VERSION ?= v1.25.3

KUBECONFIG                  ?= $(HOME)/.kube/config
LOCAL                       ?= true
TEST_MODE                   ?= true
RESTRICTED_POLICY_RESOURCES ?= policy-resource
RESTRICTED_S3_RESOURCES     ?= s3-resource
AWS_ACCOUNT_ID              ?= 123456789012
AWS_REGION                  ?= us-west-2
CLUSTER_NAME                ?= k8s_test_keiko
CLUSTER_OIDC_ISSUER_URL     ?= https://google.com/OIDC

LOCALBIN ?= $(shell pwd)/bin

ENVTEST ?= $(LOCALBIN)/setup-envtest

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

all: build

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

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

mock: ## Generate mock implementations for interfaces
	@echo "Installing mockgen..."
	go install github.com/golang/mock/mockgen@v1.6.0
	@echo "Generating mocks..."
	@cd internal/controllers && \
	PATH=$$PATH:$(GOBIN) mockgen -destination=mocks/mock_wavefrontiface.go -package=mocks github.com/keikoproj/alert-manager/pkg/wavefront Interface
	@echo "Mock generation completed"

ENVTEST_ASSETS_DIR=$(shell pwd)/bin/k8s/$(ENVTEST_K8S_VERSION)-$(OSNAME)-$(shell go env GOARCH)

.PHONY: setup-envtest
setup-envtest: $(ENVTEST) ## Download and setup the test environment binaries
	mkdir -p $(LOCALBIN)/k8s
	@echo "Setting up envtest with k8s version $(ENVTEST_K8S_VERSION)"
	KUBEBUILDER_ASSETS=$(ENVTEST_ASSETS_DIR) $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir=$(ENVTEST_ASSETS_DIR)
	@echo "Test environment binaries installed at $(ENVTEST_ASSETS_DIR)"

kind: $(KIND) ## Download kind locally if necessary.
	@echo "kind binary already exists at $(KIND)"

$(KIND): $(LOCALBIN)
	@echo "Installing kind..."
	@curl -Lo $(KIND) "https://kind.sigs.k8s.io/dl/v$(KIND_VERSION)/kind-$(OSNAME)-amd64"
	@chmod +x $(KIND)
	@echo "kind binary installed at $(KIND)"

# We're bypassing the mock/generate/manifests for the unit tests to simplify the process
unit-test: ## Run only unit tests without CRD generation
	@echo "Running unit tests only (no CRD generation)..."
	TEST_MODE=true LOCAL=true PATH=$$PATH:$(GOBIN) go test ./internal/... ./pkg/... -v

# We're creating a more focused test target that skips controllers
util-test: ## Run only utility and package tests that don't require Kubernetes
	@echo "Running utility tests only (skipping controllers)..."
	TEST_MODE=true LOCAL=true PATH=$$PATH:$(GOBIN) go test ./internal/template/... ./internal/utils/... ./pkg/wavefront/... -v

# Run with properly setup test environment using the envtest approach
envtest-test: setup-envtest fmt vet mock ## Run tests in envtest with the proper setup
	@echo "Running all tests using envtest..."
	TEST_MODE=true \
	KUBEBUILDER_ASSETS="$(ENVTEST_ASSETS_DIR)" \
	LOCAL=$(LOCAL) \
	PATH=$$PATH:$(GOBIN) go test ./... -v

test: envtest-test ## Main test target - uses envtest setup with proper binaries

##@ Build

build: mock generate fmt vet ## Build manager binary.
	go build -o bin/manager cmd/main.go

run: manifests generate fmt vet ## Run a controller from your host.
	go run cmd/main.go

docker-build: test ## Build docker image with the manager.
	docker build -t ${IMG} .

docker-push: ## Push docker image with the manager.
	docker push ${IMG}

##@ Deployment

install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl delete -f -

# Update controller-gen installation to better support ARM architectures
CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

KUSTOMIZE = $(shell pwd)/bin/kustomize
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary. If wrong version is installed, it will be removed before downloading.
$(KUSTOMIZE): $(LOCALBIN)
	@if test -x $(LOCALBIN)/kustomize && ! $(LOCALBIN)/kustomize version | grep -q $(KUSTOMIZE_VERSION); then \
		echo "$(LOCALBIN)/kustomize version is not expected $(KUSTOMIZE_VERSION). Removing it before installing."; \
		rm -rf $(LOCALBIN)/kustomize; \
	fi
	test -s $(LOCALBIN)/kustomize || GOBIN=$(LOCALBIN) GO111MODULE=on go install sigs.k8s.io/kustomize/kustomize/v3@$(KUSTOMIZE_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

##@ CI

# Integration test target for CI - combines all tests and generates coverage report
.PHONY: ci-test
ci-test: setup-envtest fmt vet mock ## Run comprehensive tests for CI pipelines
	@echo "Running comprehensive CI tests with coverage..."
	TEST_MODE=true \
	KUBEBUILDER_ASSETS="$(ENVTEST_ASSETS_DIR)" \
	LOCAL=$(LOCAL) \
	PATH=$$PATH:$(GOBIN) go test ./... -v -coverprofile cover.out

$(LOCALBIN):
	mkdir -p $(LOCALBIN)
