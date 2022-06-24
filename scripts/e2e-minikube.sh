#!/bin/bash

readonly usage="Usage: e2e-minikube.sh --test-tag <test-id>"

readonly LOCAL_REGISTRY="localhost:5000"
readonly BUILD_IMAGE="websphere-liberty-operator:latest"

readonly RUNASUSER="\n  securityContext:\n    runAsUser: 1001"
readonly APPIMAGE='applicationImage:\s'
readonly IMAGES=('k8s.gcr.io\/pause:2.0' 'navidsh\/demo-day')

# setup_env: Download kubectl cli and Minikube, start Minikube, and create a test project
setup_env() {
    # Install Minikube and Start a cluster
    echo "****** Installing and starting Minikube"
    scripts/installers/install-minikube.sh || {
        echo "Error: Failed to install minikube"
        exit 1
    }

    eval "$(minikube docker-env --profile=minikube)" && export DOCKER_CLI='docker'

    readonly TEST_NAMESPACE="wlo-test-${TEST_TAG}"

    echo "****** Creating test namespace: ${TEST_NAMESPACE}"
    kubectl create namespace "${TEST_NAMESPACE}"
    kubectl config set-context $(kubectl config current-context) --namespace="${TEST_NAMESPACE}"

    ## Create service account for Kuttl tests
    kubectl apply -f config/rbac/kind-kuttl-rbac.yaml
    
    ## Add label to node for affinity test
    kubectl label node "minikube" kuttlTest=test1

    ## Run Local Registry
    docker run -d -p 5000:5000 --restart=always --name local-registry registry
}

build_push() {
    ## Build Docker image and push to local registry
    docker build -t "${LOCAL_REGISTRY}/${BUILD_IMAGE}" .
    docker push "${LOCAL_REGISTRY}/${BUILD_IMAGE}"
}

# install_wlo: Kustomize and install WebSphere-Liberty-Operator
install_wlo() {
    echo "****** Installing WLO in namespace: ${TEST_NAMESPACE}"
    kubectl apply -f bundle/manifests/liberty.websphere.ibm.com_webspherelibertyapplications.yaml
    kubectl apply -f bundle/manifests/liberty.websphere.ibm.com_webspherelibertydumps.yaml
    kubectl apply -f bundle/manifests/liberty.websphere.ibm.com_webspherelibertytraces.yaml

    sed -i "s|image: .*|image: ${LOCAL_REGISTRY}\/${BUILD_IMAGE}|
            s|WEBSPHERE_LIBERTY_WATCH_NAMESPACE|${TEST_NAMESPACE}|" internal/deploy/kubectl/websphereliberty-app-operator.yaml

    kubectl apply -f internal/deploy/kubectl/websphereliberty-app-operator.yaml -n ${TEST_NAMESPACE}
}

install_tools() {
    echo "****** Installing Prometheus"
    kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/bundle.yaml

    echo "****** Installing Knative"
    kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.3.0/serving-crds.yaml
    kubectl apply -f https://github.com/knative/eventing/releases/download/knative-v1.3.0/eventing-crds.yaml

    echo "****** Installing Cert Manager"
    kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.8.0/cert-manager.yaml

    echo "****** Enabling Ingress"
    minikube addons enable ingress
}

## cleanup: Delete generated resources that are not bound to a test TEST_NAMESPACE.
cleanup() {
    echo
    echo "****** Cleaning up test environment..."

    ## Restore tests
    mv bundle/tests/scorecard/kuttl/ingress bundle/tests/scorecard/kind-kuttl/
    mv bundle/tests/scorecard/kuttl/ingress-certificate bundle/tests/scorecard/kind-kuttl/

    mv bundle/tests/scorecard/kind-kuttl/network-policy bundle/tests/scorecard/kuttl/
    mv bundle/tests/scorecard/kind-kuttl/network-policy-multiple-apps bundle/tests/scorecard/kuttl/
    mv bundle/tests/scorecard/kind-kuttl/routes bundle/tests/scorecard/kuttl/
    mv bundle/tests/scorecard/kind-kuttl/route-certificate bundle/tests/scorecard/kuttl/
    mv bundle/tests/scorecard/kind-kuttl/image-stream bundle/tests/scorecard/kuttl/
    mv bundle/tests/scorecard/kind-kuttl/stream bundle/tests/scorecard/kuttl/

    git checkout bundle/tests/scorecard internal/deploy

    # Clean up env
    docker rm --force local-registry
    kubectl delete namespace "${TEST_NAMESPACE}"
    minikube stop && minikube delete
}

setup_test() {
    echo "****** Installing kuttl"
    mkdir krew && cd krew
    curl -OL https://github.com/kubernetes-sigs/krew/releases/latest/download/krew-linux_amd64.tar.gz \
    && tar -xvzf krew-linux_amd64.tar.gz \
    && ./krew-linux_amd64 install krew
    cd .. && rm -rf krew
    export PATH="$HOME/.krew/bin:$PATH"
    kubectl krew install kuttl

    ## Add tests for minikube
    mv bundle/tests/scorecard/kind-kuttl/ingress bundle/tests/scorecard/kuttl/
    mv bundle/tests/scorecard/kind-kuttl/ingress-certificate bundle/tests/scorecard/kuttl/
    
    ## Remove tests that do not apply for minikube
    mv bundle/tests/scorecard/kuttl/network-policy bundle/tests/scorecard/kind-kuttl/
    mv bundle/tests/scorecard/kuttl/network-policy-multiple-apps bundle/tests/scorecard/kind-kuttl/
    mv bundle/tests/scorecard/kuttl/routes bundle/tests/scorecard/kind-kuttl/
    mv bundle/tests/scorecard/kuttl/route-certificate bundle/tests/scorecard/kind-kuttl/
    mv bundle/tests/scorecard/kuttl/image-stream bundle/tests/scorecard/kind-kuttl/

    for image in "${IMAGES[@]}"; do
        files=($(grep -rwl 'bundle/tests/scorecard/kuttl/' -e $APPIMAGE$image))
        for file in "${files[@]}"; do
            sed -i "s/$image/$image$RUNASUSER/" $file
        done
    done
}

main() {
    parse_args "$@"
     
    if [[ -z "${TEST_TAG}" ]]; then
        echo "****** Missing test id, see usage"
        echo "${usage}"
        exit 1
    fi

    echo "****** Setting up test environment..."
    setup_env
    build_push
    install_wlo
    install_tools

    # Wait for operator deployment to be ready
    while [[ $(kubectl get deploy wlo-controller-manager -o jsonpath='{ .status.readyReplicas }') -ne "1" ]]; do
        echo "****** Waiting for wlo-controller-manager to be ready..."
        sleep 10
    done
    echo "****** wlo-controller-manager deployment is ready..."
    
    setup_test
    
    echo "****** Starting minikube scorecard tests..."
    operator-sdk scorecard --verbose --selector=suite=kuttlsuite --namespace "${TEST_NAMESPACE}" --service-account scorecard-kuttl --wait-time 30m ./bundle || {
        echo "****** Scorecard tests failed..."
        echo "****** Cleaning up test environment..."
        cleanup
    }
    return $?
}

parse_args() {
    while [ $# -gt 0 ]; do
        case "$1" in
        --test-tag)
            shift
            readonly TEST_TAG="${1}"
            ;;
        *)
            echo "Error: Invalid argument - $1"
            echo "$usage"
            exit 1
            ;;
        esac
        shift
    done
}

# Always do cleanup when script exits.
# TODO: Add flag to keep enviornment in case we need to debug test failures.
trap cleanup EXIT

main "$@"
