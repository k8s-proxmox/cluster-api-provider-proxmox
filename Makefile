
# Image URL to use all building/pushing image targets
REGISTRY := ghcr.io
PROJECT := k8s-proxmox/cluster-api-provider-proxmox
RELEASE_TAG := latest
IMG ?= $(REGISTRY)/$(PROJECT):$(RELEASE_TAG)
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.26.1

# add localbin to PATH
LOCALBIN ?= $(shell pwd)/bin
export PATH := $(abspath $(LOCALBIN)):$(PATH)

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

GOARCH  := $(shell go env GOARCH)
GOOS    := $(shell go env GOOS)

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
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

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: goimports ## Run go fmt against code.
	go fmt ./...
	$(GOIMPORTS) -w ./

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: lint
lint: golangci-lint ## Run golangci-lint
	$(GOLANGCI_LINT) run

CLUSTER_NAME := cappx-test

.PHONY: create-workload-cluster
create-workload-cluster: $(KUSTOMIZE) $(ENVSUBST) $(KUBECTL)
	export CLUSTER_NAME=$(CLUSTER_NAME) && $(KUSTOMIZE) build templates | $(ENVSUBST) | $(KUBECTL) apply -f -

.PHONY: delete-workload-cluster
delete-workload-cluster: $(KUBECTL)
	$(KUBECTL) delete cluster $(CLUSTER_NAME)

##@ Testing

SETUP_ENVTEST_VER := v0.0.0-20211110210527-619e6b92dab9
SETUP_ENVTEST := $(LOCALBIN)/setup-envtest
GINKGO_TIMEOUT ?= 30m

.PHONY: test
test: generate manifests fmt $(SETUP_ENVTEST)  ## Run unit and integration test
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./api/... ./cloud/... ./controllers/... $(TEST_ARGS)

.PHONY: unit-test  ## Run unit tests
unit-test: generate manifests fmt $(SETUP_ENVTEST)
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./api/... ./cloud/... ./controllers/... --ginkgo.label-filter=unit $(TEST_ARGS)

.PHONY: test-cover
test-cover: ## Run unit and integration tests and generate coverage report
	$(MAKE) test TEST_ARGS="$(TEST_ARGS) -coverprofile=coverage.out"
	go tool cover -func=coverage.out -o coverage.txt
	go tool cover -html=coverage.out -o coverage.html

.PHONY: unit-test-cover
unit-test-cover: ## Run unit tests and generate coverage report
	$(MAKE) unit-test TEST_ARGS="$(TEST_ARGS) -coverprofile=coverage.out"
	go tool cover -func=coverage.out -o coverage.txt
	go tool cover -html=coverage.out -o coverage.html

E2E_DIR = $(shell pwd)/internal/test/e2e
E2E_IMG := ghcr.io/k8s-proxmox/cluster-api-provider-proxmox:e2e
.PHONY: generate-e2e-templates
generate-e2e-templates: $(KUSTOMIZE) ## Generate cluster-templates for e2e
	cp templates/cluster-template* $(E2E_DIR)/data/infrastructure-proxmox/templates
	$(KUSTOMIZE) build $(E2E_DIR)/data/infrastructure-proxmox/templates > $(E2E_DIR)/data/infrastructure-proxmox/main/cluster-template.yaml

.PHONY: build-e2e-image
build-e2e-image: ## Build cappx image to be used for e2e test
	IMG=${E2E_IMG} $(MAKE) docker-build

USE_EXISTING_CLUSTER := false
.PHONY: e2e
e2e: generate-e2e-templates build-e2e-image cleanup-e2e-artifacts $(KUBECTL) ## Run e2e test
	go test $(E2E_DIR)/... -v \
	-timeout=$(GINKGO_TIMEOUT) \
	--e2e.artifacts-folder=$(E2E_DIR) \
	--e2e.use-existing-cluster="$(USE_EXISTING_CLUSTER)"

.PHONY: cleanup-e2e-artifacts
cleanup-e2e-artifacts: ## delete some e2e artifacts 
	rm -rf $(E2E_DIR)/clusters
	rm -rf $(E2E_DIR)/kind

##@ Build

.PHONY: build
build: manifests generate fmt vet ## Build manager binary.
	go build -o bin/manager cmd/main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./cmd/main.go

# If you wish built the manager image targeting other platforms you can use the --platform flag.
# (i.e. docker build --platform linux/arm64 ). However, you must enable docker buildKit for it.
# More info: https://docs.docker.com/develop/develop-images/build_enhancements/
.PHONY: docker-build
docker-build: unit-test ## Build docker image with the manager.
	docker build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${IMG}

# PLATFORMS defines the target platforms for  the manager image be build to provide support to multiple
# architectures. (i.e. make docker-buildx IMG=myregistry/mypoperator:0.0.1). To use this option you need to:
# - able to use docker buildx . More info: https://docs.docker.com/build/buildx/
# - have enable BuildKit, More info: https://docs.docker.com/develop/develop-images/build_enhancements/
# - be able to push the image for your registry (i.e. if you do not inform a valid value via IMG=<myregistry/image:<tag>> then the export will fail)
# To properly provided solutions that supports more than one platform you should use this option.
PLATFORMS ?= linux/arm64,linux/amd64 #,linux/s390x,linux/ppc64le
.PHONY: docker-buildx
docker-buildx: ## Build and push docker image for the manager for cross-platform support
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- docker buildx create --name project-v3-builder
	docker buildx use project-v3-builder
	- docker buildx build --push --platform=$(PLATFORMS) --tag ${IMG} -f Dockerfile.cross .
	- docker buildx rm project-v3-builder
	rm Dockerfile.cross

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

##@ Release

## Location to output for release
RELEASE_DIR := $(shell pwd)/out
$(RELEASE_DIR):
	mkdir -p $(RELEASE_DIR)

RELEASE_TAG := $(shell git describe --abbrev=0 --tags)

.PHONY: release
release: ## Builds all the manifests/config files to publish with a release
	@echo "Building assets for Release \"$(RELEASE_TAG)\""
	$(MAKE) release-manifests
	$(MAKE) release-metadata
	$(MAKE) release-templates

.PHONY: release-manifests
release-manifests: $(KUSTOMIZE) $(RELEASE_DIR) ## Builds the manifests to publish with a release
	cd config/manager && $(KUSTOMIZE) edit set image controller=${REGISTRY}/${PROJECT}:${RELEASE_TAG}
	$(KUSTOMIZE) build config/default > $(RELEASE_DIR)/infrastructure-components.yaml

.PHONY: release-metadata
release-metadata: $(RELEASE_DIR)
	cp metadata.yaml $(RELEASE_DIR)/metadata.yaml

.PHONY: release-templates
release-templates: $(RELEASE_DIR)
	cp templates/cluster-template* $(RELEASE_DIR)/


##@ Build Dependencies

## Location to install dependencies to
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
ENVSUBST ?= $(LOCALBIN)/envsubst
KUBECTL ?= $(LOCALBIN)/kubectl
GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint
GOIMPORTS ?= $(LOCALBIN)/goimports

## Tool Versions
KUSTOMIZE_VERSION ?= v5.0.0
CONTROLLER_TOOLS_VERSION ?= v0.11.3
ENVSUBST_VER ?= v1.4.2
KUBECTL_VER := v1.25.10

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary. If wrong version is installed, it will be removed before downloading.
$(KUSTOMIZE): $(LOCALBIN)
	@if test -x $(LOCALBIN)/kustomize && ! $(LOCALBIN)/kustomize version | grep -q $(KUSTOMIZE_VERSION); then \
		echo "$(LOCALBIN)/kustomize version is not expected $(KUSTOMIZE_VERSION). Removing it before installing."; \
		rm -rf $(LOCALBIN)/kustomize; \
	fi
	test -s $(LOCALBIN)/kustomize || { curl -Ss $(KUSTOMIZE_INSTALL_SCRIPT) --output install_kustomize.sh && bash install_kustomize.sh $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); rm install_kustomize.sh; }

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: envsubst
envsubst: $(ENVSUBST)
$(ENVSUBST): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install github.com/a8m/envsubst/cmd/envsubst@$(ENVSUBST_VER)

.PHONY: kubectl
kubectl: $(KUBECTL)
$(KUBECTL): $(LOCALBIN)
	curl --retry 3 -fsL https://dl.k8s.io/release/$(KUBECTL_VER)/bin/$(GOOS)/$(GOARCH)/kubectl -o $(LOCALBIN)/kubectl
	chmod +x $(KUBECTL)

.PHONY: setup-envtest
setup-envtest: $(SETUP_ENVTEST)
$(SETUP_ENVTEST): go.mod # Build setup-envtest from tools folder.
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@$(SETUP_ENVTEST_VER)

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT)
$(GOLANGCI_LINT): $(LOCALBIN)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(LOCALBIN) v1.54.0

.PHONY: goimports
goimports: $(GOIMPORTS)
$(GOIMPORTS): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install golang.org/x/tools/cmd/goimports@latest
