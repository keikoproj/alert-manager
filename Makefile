
# Image URL to use all building/pushing image targets
IMG ?=  keikoproj/alert-manager:latest

# Tools required to run the full suite of tests properly
OSNAME           ?= $(shell uname -s | tr A-Z a-z)
KUBEBUILDER_VER  ?= 3.0.0
KUBEBUILDER_ARCH ?= amd64
ENVTEST_K8S_VERSION = 1.28.0

KUBECONFIG                  ?= $(HOME)/.kube/config
LOCAL                       ?= true
RESTRICTED_POLICY_RESOURCES ?= policy-resource
RESTRICTED_S3_RESOURCES     ?= s3-resource
AWS_ACCOUNT_ID              ?= 123456789012
AWS_REGION                  ?= us-west-2
CLUSTER_NAME                ?= k8s_test_keiko
CLUSTER_OIDC_ISSUER_URL     ?= https://google.com/OIDC

LOCALBIN ?= $(shell pwd)/bin
# Export local bin to path for all recipes
export PATH := $(LOCALBIN):$(PATH)

ENVTEST ?= $(LOCALBIN)/setup-envtest

## Tool Binaries
MOCKGEN ?= $(LOCALBIN)/mockgen
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen

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
	$(CONTROLLER_GEN) crd rbac:roleName=role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

mock: $(MOCKGEN)
	@echo "mockgen is in progess"
	@for pkg in $(shell go list ./...) ; do \
		go generate ./... ;\
	done

ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
test: mock manifests generate fmt vet envtest ## Run tests.
	KUBECONFIG=$(KUBECONFIG) \
	LOCAL=$(LOCAL) \
	RESTRICTED_POLICY_RESOURCES=$(RESTRICTED_POLICY_RESOURCES) \
	RESTRICTED_S3_RESOURCES=$(RESTRICTED_S3_RESOURCES) \
	AWS_ACCOUNT_ID=$(AWS_ACCOUNT_ID) \
	AWS_REGION=$(AWS_REGION) \
	CLUSTER_NAME=$(CLUSTER_NAME) \
	CLUSTER_OIDC_ISSUER_URL="$(CLUSTER_OIDC_ISSUER_URL)" \
	DEFAULT_TRUST_POLICY=$(DEFAULT_TRUST_POLICY) \
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -coverprofile cover.out

##@ Build

.PHONY: build
build: $(LOCALBIN)/manager
$(LOCALBIN)/manager: mock generate fmt vet ## Build manager binary.
	go build -o $(LOCALBIN)/manager cmd/main.go

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


## Tool Versions
MOCKGEN_VERSION ?= v1.6.0
KUSTOMIZE_VERSION ?= v4.2.0
CONTROLLER_TOOLS_VERSION ?= v0.17.0

$(MOCKGEN): $(LOCALBIN) ## Download mockgen if necessary.
	GOBIN=$(LOCALBIN) go install github.com/golang/mock/mockgen@$(MOCKGEN_VERSION)

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN)

# go-install-tool will 'go install' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-install-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

$(LOCALBIN):
	mkdir -p $(LOCALBIN)