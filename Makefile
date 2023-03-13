# VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g make bundle VERSION=0.0.2)
# - use environment variables to overwrite this value (e.g export VERSION=0.0.2)
VERSION ?= 1.1.0
OPERATOR_SDK_RELEASE_VERSION ?= v1.24.0

# CHANNELS define the bundle channels used in the bundle.
# Add a new line here if you would like to change its default config. (E.g CHANNELS = "preview,fast,stable")
# To re-generate a bundle for other specific channels without changing the standard setup, you can:
# - use the CHANNELS as arg of the bundle target (e.g make bundle CHANNELS=preview,fast,stable)
# - use environment variables to overwrite this value (e.g export CHANNELS="preview,fast,stable")
CHANNELS ?= v1.1
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif

# DEFAULT_CHANNEL defines the default channel used in the bundle.
# Add a new line here if you would like to change its default config. (E.g DEFAULT_CHANNEL = "stable")
# To re-generate a bundle for any other default channel without changing the default setup, you can:
# - use the DEFAULT_CHANNEL as arg of the bundle target (e.g make bundle DEFAULT_CHANNEL=stable)
# - use environment variables to overwrite this value (e.g export DEFAULT_CHANNEL="stable")
DEFAULT_CHANNEL ?= v1.1
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# IMAGE_TAG_BASE defines the docker.io namespace and part of the image name for remote images.
# This variable is used to construct full image tags for bundle and catalog images.
#
# For example, running 'make bundle-build bundle-push catalog-build catalog-push' will build and push both
# openliberty.io/op-test-bundle:$VERSION and openliberty.io/op-test-catalog:$VERSION.
IMAGE_TAG_BASE ?= icr.io/cpopen/websphere-liberty-operator

# OPERATOR_IMAGE defines the docker.io namespace and part of the image name for remote images.
OPERATOR_IMAGE ?= icr.io/cpopen/websphere-liberty-operator

# BUNDLE_IMG defines the image:tag used for the bundle.
# You can use it as an arg. (E.g make bundle-build BUNDLE_IMG=<some-registry>/<project-name-bundle>:<tag>)
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-bundle:daily

# A comma-separated list of bundle images (e.g. make catalog-build BUNDLE_IMGS=example.com/operator-bundle:v0.1.0,example.com/operator-bundle:v0.2.0).
# These images MUST exist in a registry and be pull-able.
BUNDLE_IMGS ?= $(BUNDLE_IMG)

# BUNDLE_GEN_FLAGS are the flags passed to the operator-sdk generate bundle command
BUNDLE_GEN_FLAGS ?= -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)

# USE_IMAGE_DIGESTS defines if images are resolved via tags or digests
# You can enable this value if you would like to use SHA Based Digests
# To enable set flag to true
USE_IMAGE_DIGESTS ?= false
ifeq ($(USE_IMAGE_DIGESTS), true)
    BUNDLE_GEN_FLAGS += --use-image-digests
endif

# Image URL to use all building/pushing image targets
IMG ?= icr.io/cpopen/websphere-liberty-operator:daily

# The image tag given to the resulting catalog image (e.g. make catalog-build CATALOG_IMG=example.com/operator-catalog:v0.2.0).
CATALOG_IMG ?= $(IMAGE_TAG_BASE)-catalog:v$(VERSION)

# Set CATALOG_BASE_IMG to an existing catalog image tag to add $BUNDLE_IMGS to that image.
ifneq ($(origin CATALOG_BASE_IMG), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
endif

PUBLISH_REGISTRY=docker.io

# Type of release. Can be "daily", "releases", or a release tag.
RELEASE_TARGET := $(or ${RELEASE_TARGET}, ${TRAVIS_TAG}, daily)

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

CREATEDAT ?= AUTO
ifeq ($(CREATEDAT), AUTO)
CREATEDAT := $(shell date +%Y-%m-%dT%TZ)
endif

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:crdVersions=v1,generateEmbeddedObjectMeta=true"

# Produce files under deploy/kustomize/daily with default namespace
KUSTOMIZE_NAMESPACE = default
KUSTOMIZE_IMG = cp.stg.icr.io/cp/websphere-liberty-operator:main

# Use docker if available. Otherwise default to podman. 
# Override choice by setting CONTAINER_COMMAND
CHECK_DOCKER_RC=$(shell docker -v > /dev/null 2>&1; echo $$?)
ifneq (0, $(CHECK_DOCKER_RC))
CONTAINER_COMMAND ?= podman
# Setup parameters for TLS verify, default if unspecified is true
ifeq (false, $(TLS_VERIFY))
PODMAN_SKIP_TLS_VERIFY="--tls-verify=false"
SKIP_TLS_VERIFY=--skip-tls
else
TLS_VERIFY ?= true
PODMAN_SKIP_TLS_VERIFY="--tls-verify=true"
endif
else
CONTAINER_COMMAND ?= docker
endif

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

##@ Setup

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
# find or download controller-gen
# download controller-gen if necessary
.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.9.2

KUSTOMIZE ?= $(LOCALBIN)/kustomize
KUSTOMIZE_VERSION ?= 3.8.7
KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/release-kustomize-v3.8/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s $(KUSTOMIZE_VERSION) $(LOCALBIN)


.PHONY: setup
setup: ## Ensure Operator SDK is installed.
	./scripts/installers/install-operator-sdk.sh ${OPERATOR_SDK_RELEASE_VERSION}

.PHONY: setup-manifest
setup-manifest: ## Install manifest tool.
	./scripts/installers/install-manifest-tool.sh

# Install Podman.
install-podman:
	./scripts/installers/install-podman.sh

# Install OPM.
install-opm:
	./scripts/installers/install-opm.sh

##@ Development

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: bundle
bundle: manifests setup kustomize ## Generate bundle manifests and metadata, then validate generated files.
	scripts/update-sample.sh
	sed -i.bak "s,IMAGE,${IMG},g;s,CREATEDAT,${CREATEDAT},g" config/manifests/patches/csvAnnotations.yaml
	operator-sdk generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle $(BUNDLE_GEN_FLAGS)
	./scripts/csv_description_update.sh update_csv

	$(KUSTOMIZE) build config/kustomize/crd -o internal/deploy/kustomize/daily/base/websphere-liberty-crd.yaml
	cd config/kustomize/operator && $(KUSTOMIZE) edit set namespace $(KUSTOMIZE_NAMESPACE)
	$(KUSTOMIZE) build config/kustomize/operator -o internal/deploy/kustomize/daily/base/websphere-liberty-deployment.yaml
	sed -i.bak "s,${IMG},${KUSTOMIZE_IMG},g;s,serviceAccountName: controller-manager,serviceAccountName: websphere-liberty-controller-manager,g" internal/deploy/kustomize/daily/base/websphere-liberty-deployment.yaml
	$(KUSTOMIZE) build config/kustomize/roles -o internal/deploy/kustomize/daily/base/websphere-liberty-roles.yaml

	mv config/manifests/patches/csvAnnotations.yaml.bak config/manifests/patches/csvAnnotations.yaml
	rm internal/deploy/kustomize/daily/base/websphere-liberty-deployment.yaml.bak
	operator-sdk bundle validate ./bundle

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
.PHONY: test
test: manifests generate fmt vet ## Run tests.
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.7.2/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test ./... -coverprofile cover.out

.PHONY: unit-test
unit-test: ## Run unit tests
	go test -v -mod=vendor -tags=unit github.com/WASdev/websphere-liberty-operator/...

.PHONY: run
run: manifests generate fmt vet ## Run a controller against the configured Kubernetes cluster in ~/.kube/config from your host.
	go run ./main.go

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

##@ Build

.PHONY: build
build: generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go

.PHONY: docker-login
docker-login:
	docker login -u "${DOCKER_USERNAME}" -p "${DOCKER_PASSWORD}"

.PHONY: docker-build
docker-build: test ## Build docker image with the manager.
	$(CONTAINER_COMMAND) build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	$(CONTAINER_COMMAND) push $(PODMAN_SKIP_TLS_VERIFY) ${IMG}

.PHONY: bundle-build
bundle-build: ## Build the bundle image.
	$(CONTAINER_COMMAND) build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: bundle-push
bundle-push: ## Push the bundle image.
	$(CONTAINER_COMMAND) push $(PODMAN_SKIP_TLS_VERIFY) $(BUNDLE_IMG)

build-pipeline-releases:
	./scripts/build-releases.sh -u "${PIPELINE_USERNAME}" -p "${PIPELINE_PASSWORD}" --registry "${PIPELINE_REGISTRY}" --image "${PIPELINE_REGISTRY}/${PIPELINE_OPERATOR_IMAGE}"	--target "${RELEASE_TARGET}"

build-artifactory-releases:
	./scripts/build-releases.sh -u "${ARTIFACTORY_USERNAME}" -p "${ARTIFACTORY_TOKEN}" --registry "${ARTIFACTORY_REPO_URL}" --image "${ARTIFACTORY_REPO_URL}/${PIPELINE_OPERATOR_IMAGE}"	--target "${RELEASE_TARGET}"

build-all-releases: build-pipeline-releases build-artifactory-releases


# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

.PHONY: opm
OPM = ./bin/opm
opm: ## Download opm locally if necessary.
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.15.1/$${OS}-$${ARCH}-opm ;\
	chmod +x $(OPM) ;\
	}
else
OPM = $(shell which opm)
endif
endif

# Build a catalog image by adding bundle images to an empty catalog using the operator package manager tool, 'opm'.
# This recipe invokes 'opm' in 'semver' bundle add mode. For more information on add modes, see:
# https://github.com/operator-framework/community-operators/blob/7f1438c/docs/packaging-operator.md#updating-your-existing-operator
.PHONY: catalog-build
catalog-build: opm ## Build a catalog image.
	$(OPM) index add $(SKIP_TLS_VERIFY) --container-tool $(CONTAINER_COMMAND)  --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) $(FROM_INDEX_OPT) --permissive

# Push the catalog image.
.PHONY: catalog-push
catalog-push: ## Push a catalog image.
	$(MAKE) docker-push IMG=$(CATALOG_IMG)

build-manifest: setup-manifest
	./scripts/build-manifest.sh --image "${PUBLISH_REGISTRY}/${OPERATOR_IMAGE}" --target "${RELEASE_TARGET}"

kind-e2e-test:
	./scripts/e2e-kind.sh --test-tag "${TRAVIS_BUILD_NUMBER}"

build-pipeline-manifest: setup-manifest
	./scripts/build-manifest.sh -u "${PIPELINE_USERNAME}" -p "${PIPELINE_PASSWORD}" --registry "${PIPELINE_REGISTRY}" --image "${PIPELINE_REGISTRY}/${PIPELINE_OPERATOR_IMAGE}"	--target "${RELEASE_TARGET}"

build-artifactory-manifest: setup-manifest
	./scripts/build-manifest.sh -u "${ARTIFACTORY_USERNAME}" -p "${ARTIFACTORY_TOKEN}" --registry "${ARTIFACTORY_REPO_URL}" --image "${ARTIFACTORY_REPO_URL}/${PIPELINE_OPERATOR_IMAGE}"	--target "${RELEASE_TARGET}"

build-all-manifest: build-pipeline-manifest build-artifactory-manifest

bundle-pipeline:
	./scripts/bundle-release.sh -u "${PIPELINE_USERNAME}" -p "${PIPELINE_PASSWORD}" --registry "${PIPELINE_REGISTRY}" --prod-image "${PIPELINE_PRODUCTION_IMAGE}" --image "${PIPELINE_REGISTRY}/${PIPELINE_OPERATOR_IMAGE}" --release "${RELEASE_TARGET}"

bundle-artifactory:
	./scripts/bundle-release.sh -u "${ARTIFACTORY_USERNAME}" -p "${ARTIFACTORY_TOKEN}" --registry "${ARTIFACTORY_REPO_URL}" --prod-image "${PIPELINE_PRODUCTION_IMAGE}" --image "${ARTIFACTORY_REPO_URL}/${PIPELINE_OPERATOR_IMAGE}" --release "${RELEASE_TARGET}"

bundle-all: bundle-pipeline bundle-artifactory

catalog-pipeline-build: opm ## Build a catalog image.
	./scripts/catalog-build.sh -n "v${OPM_VERSION}" -b "${REDHAT_BASE_IMAGE}" -o "${OPM}" --container-tool "docker" -i "${PIPELINE_REGISTRY}/${PIPELINE_OPERATOR_IMAGE}-bundle:${RELEASE_TARGET}" -p "${PIPELINE_PRODUCTION_IMAGE}-bundle" -a "${PIPELINE_REGISTRY}/${PIPELINE_OPERATOR_IMAGE}-catalog:${RELEASE_TARGET}" -t "${PWD}/operator-build" -v "${VERSION}"

catalog-artifactory-build: opm ## Build a catalog image.
	./scripts/catalog-build.sh -n "v${OPM_VERSION}" -b "${REDHAT_BASE_IMAGE}" -o "${OPM}" --container-tool "docker" -i "${ARTIFACTORY_REPO_URL}/${PIPELINE_OPERATOR_IMAGE}-bundle:${RELEASE_TARGET}" -p "${PIPELINE_PRODUCTION_IMAGE}-bundle" -a "${ARTIFACTORY_REPO_URL}/${PIPELINE_OPERATOR_IMAGE}-catalog:${RELEASE_TARGET}" -t "${PWD}/operator-build"

catalog-all-build: opm catalog-pipeline-build catalog-artifactory-build ## Build a catalog image

catalog-pipeline-push: ## Push a catalog image.
	$(MAKE) docker-push IMG="${PIPELINE_REGISTRY}/${PIPELINE_OPERATOR_IMAGE}-catalog:${RELEASE_TARGET}"

catalog-artifactory-push: ## Push a catalog image.
	$(MAKE) docker-push IMG="${ARTIFACTORY_REPO_URL}/${PIPELINE_OPERATOR_IMAGE}-catalog:${RELEASE_TARGET}"

catalog-all-push: catalog-pipeline-push catalog-artifactory-push ## Push a catalog image.

test-e2e:
	./scripts/e2e-release.sh --registry-name default-route --registry-namespace openshift-image-registry \
                     --test-tag "${TRAVIS_BUILD_NUMBER}" --target "${RELEASE_TARGET}"

test-pipeline-e2e:
	./scripts/pipeline/ocp-cluster-e2e.sh -u "${DOCKER_USERNAME}" -p "${DOCKER_PASSWORD}" \
                     --cluster-url "${CLUSTER_URL}" --cluster-user "${CLUSTER_USER}" --cluster-token "${CLUSTER_TOKEN}" \
                     --registry-name "${PIPELINE_REGISTRY}" --registry-image "${PIPELINE_OPERATOR_IMAGE}" \
                     --registry-user "${PIPELINE_USERNAME}" --registry-password "${PIPELINE_PASSWORD}" \
                     --test-tag "${TRAVIS_BUILD_NUMBER}" --release "${RELEASE_TARGET}" --channel "${DEFAULT_CHANNEL}"

build-releases:
	./scripts/build-releases.sh --image "${PUBLISH_REGISTRY}/${OPERATOR_IMAGE}" --target "${RELEASE_TARGET}"

bundle-releases:
	./scripts/bundle-releases.sh --image "${PUBLISH_REGISTRY}/${OPERATOR_IMAGE}" --target "${RELEASE_TARGET}"

bundle-build-podman:
	podman build -f bundle.Dockerfile -t "${BUNDLE_IMG}"

bundle-push-podman:
	podman push --format=docker "${BUNDLE_IMG}"

build-catalog:
	opm index add --bundles "${BUNDLE_IMG}" --tag "${CATALOG_IMG}"

push-catalog: docker-login
	podman push --format=docker "${CATALOG_IMG}"

dev: 
	./scripts/dev.sh all -channel ${DEFAULT_CHANNEL}
