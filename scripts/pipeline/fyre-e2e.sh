#!/bin/bash

readonly usage="Usage: fyre-e2e.sh -u <docker-username> -p <docker-password> --cluster-url <url> --cluster-token <token> --registry-name <name> --registry-image <ns/image> --registry-user <user> --registry-password <password> --release <daily|release-tag> --test-tag <test-id> --catalog-image <catalog-image> --channel <channel>"
readonly OC_CLIENT_VERSION="4.6.0"
readonly CONTROLLER_MANAGER_NAME="wlo-controller-manager"

# setup_env: Download oc cli, log into our persistent cluster, and create a test project
setup_env() {
    echo "****** Installing OC CLI..."
    # Install kubectl and oc
    curl -L https://mirror.openshift.com/pub/openshift-v4/clients/ocp/${OC_CLIENT_VERSION}/openshift-client-linux.tar.gz | tar xvz
    sudo mv oc kubectl /usr/local/bin/

    # Start a cluster and login
    echo "****** Logging into remote cluster..."
    oc login "${CLUSTER_URL}" -u "${CLUSTER_USER:-kubeadmin}" -p "${CLUSTER_TOKEN}" --insecure-skip-tls-verify=true

    # Set variables for rest of script to use
    readonly TEST_NAMESPACE="wlo-test-${TEST_TAG}"

    echo "****** Creating test namespace: ${TEST_NAMESPACE} for release ${RELEASE}"
    oc new-project "${TEST_NAMESPACE}" || oc project "${TEST_NAMESPACE}"
}

## cleanup_env : Delete generated resources that are not bound to a test TEST_NAMESPACE.
cleanup_env() {
  oc delete project "${TEST_NAMESPACE}"
}

## trap_cleanup : Call cleanup_env and exit. For use by a trap to detect if the script is exited at any point.
trap_cleanup() {
  last_status=$?
  if [[ $last_status != 0 ]]; then
    cleanup_env
  fi
  exit $last_status
}

#push_images() {
#    echo "****** Logging into private registry..."
#    oc sa get-token "${SERVICE_ACCOUNT}" -n default | docker login -u unused --password-stdin "${DEFAULT_REGISTRY}" || {
#        echo "Failed to log into docker registry as ${SERVICE_ACCOUNT}, exiting..."
#        exit 1
#    }

#    echo "****** Creating pull secret using Docker config..."
#    oc create secret generic regcred --from-file=.dockerconfigjson="${HOME}/.docker/config.json" --type=kubernetes.io/dockerconfigjson

#    docker push "${BUILD_IMAGE}" || {
#        echo "Failed to push ref: ${BUILD_IMAGE} to docker registry, exiting..."
#        exit 1
#    }

#    docker push "${BUNDLE_IMAGE}" || {
#        echo "Failed to push ref: ${BUNDLE_IMAGE} to docker registry, exiting..."
#        exit 1
#    }
#}

main() {
    parse_args "$@"

    if [[ -z "${RELEASE}" ]]; then
        echo "****** Missing release, see usage"
    fi

    if [[ -z "${DOCKER_USERNAME}" || -z "${DOCKER_PASSWORD}" ]]; then
        echo "****** Missing docker authentication information, see usage"
        echo "${usage}"
        exit 1
    fi

    if [[ -z "${CLUSTER_URL}" ]] || [[ -z "${CLUSTER_TOKEN}" ]]; then
        echo "****** Missing OCP URL or token, see usage"
        echo "${usage}"
        exit 1
    fi

    if [[ -z "${REGISTRY_NAME}" ]]; then
        echo "****** Missing OCP registry name, see usage"
        echo "${usage}"
        exit 1
    fi

    if [[ -z "${REGISTRY_IMAGE}" ]]; then
        echo "****** Missing REGISTRY_IMAGE definition, see usage"
        echo "${usage}"
        exit 1
    fi

    if [[ -z "${REGISTRY_USER}" ]] || [[ -z "${REGISTRY_PASSWORD}" ]]; then
        echo "****** Missing registry authentication information, see usage"
        echo "${usage}"
        exit 1
    fi

    if [[ -z "${TEST_TAG}" ]]; then
        echo "****** Missing test tag, see usage"
        echo "${usage}"
        exit 1
    fi

    if [[ -z "${CATALOG_IMAGE}" ]]; then
        echo "****** Missing catalog image, see usage"
        echo "${usage}"
        exit 1
    fi

    if [[ -z "${CHANNEL}" ]]; then
        echo "****** Missing channel, see usage"
        echo "${usage}"
        exit 1
    fi

    echo "****** Setting up test environment..."
    setup_env

    if [[ -z "${DEBUG_FAILURE}" ]]; then
        trap trap_cleanup EXIT
    else
        echo "#####################################################################################"
        echo "WARNING: --debug-failure is set. If e2e tests fail, any created resources will remain"
        echo "on the cluster for debugging/troubleshooting. YOU MUST DELETE THESE RESOURCES when"
        echo "you're done, or else they will cause future tests to fail. To cleanup manually, just"
        echo "delete the namespace \"${TEST_NAMESPACE}\": oc delete project \"${TEST_NAMESPACE}\" "
        echo "#####################################################################################"
    fi

    # login to docker to avoid rate limiting during build
    echo "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME}" --password-stdin

    trap "rm -f /tmp/pull-secret-*.yaml" EXIT

    echo "****** Logging into private registry..."
    echo "${REGISTRY_PASSWORD}" | docker login ${REGISTRY_NAME} -u "${REGISTRY_USER}" --password-stdin

    echo "****** Creating pull secret..."
    oc create secret docker-registry regcred --docker-server=${REGISTRY_NAME} "--docker-username=${REGISTRY_USER}" "--docker-password=${REGISTRY_PASSWORD}" --docker-email=unused 

    oc get secret/regcred -o jsonpath='{.data.\.dockerconfigjson}' | base64 --decode > /tmp/pull-secret-new.yaml
    oc get secret/pull-secret -n openshift-config -o jsonpath='{.data.\.dockerconfigjson}' | base64 --decode > /tmp/pull-secret-global.yaml

    jq -s '.[1] * .[0]' /tmp/pull-secret-new.yaml /tmp/pull-secret-global.yaml > /tmp/pull-secret-merged.yaml

    echo "Updating global pull secret"
    oc set data secret/pull-secret -n openshift-config --from-file=.dockerconfigjson=/tmp/pull-secret-merged.yaml

    # Run Kuttl scorecard tests
    run_kuttl_tests "OwnNamespace"
    run_kuttl_tests "AllNamespaces"
    run_kuttl_tests "SingleNamespace"

    echo "****** Cleaning up test environment..."
    cleanup_env

    return $result
}

install_operator_and_run_scorecard_tests() {
    if [ "$INSTALL_MODE" == "AllNamespaces" ]; then
        CONTROLLER_MANAGER_NAMESPACE="openshift-operators"
        SERVICE_ACCOUNT_NAME="scorecard-kuttl-cluster-wide"
        OPERATOR_GROUP_TARGET_NAMESPACE="openshift-operators" # used for CSV cleanup
    elif [ "$INSTALL_MODE" == "SingleNamespace" ]; then
        CONTROLLER_MANAGER_NAMESPACE="openshift-marketplace"
        SERVICE_ACCOUNT_NAME="scorecard-kuttl-cluster-wide"
	OPERATOR_GROUP_NAMESPACE="openshift-marketplace"
        OPERATOR_GROUP_TARGET_NAMESPACE="${TEST_NAMESPACE}"
    elif [ "$INSTALL_MODE" == "OwnNamespace" ]; then
	CONTROLLER_MANAGER_NAMESPACE="${TEST_NAMESPACE}"
	SERVICE_ACCOUNT_NAME="scorecard-kuttl"
	OPERATOR_GROUP_NAMESPACE="${TEST_NAMESPACE}"
        OPERATOR_GROUP_TARGET_NAMESPACE="${TEST_NAMESPACE}"
    fi

    # Delete subscriptions that may be blocking the install
    oc delete subscription.operators.coreos.com websphere-liberty-operator-subscription -n ${CONTROLLER_MANAGER_NAMESPACE}

    install_operator

    # Wait for operator deployment to be ready
    while [[ $(oc get deploy "${CONTROLLER_MANAGER_NAME}" -n ${CONTROLLER_MANAGER_NAMESPACE} -o jsonpath='{ .status.readyReplicas }') -ne "1" ]]; do
        echo "****** Waiting for ${CONTROLLER_MANAGER_NAME} in namespace ${CONTROLLER_MANAGER_NAMESPACE} to be ready..."
        sleep 10
    done

    echo "****** ${CONTROLLER_MANAGER_NAME} deployment is ready..."
 
    echo "****** Starting scorecard tests..."
    operator-sdk scorecard --verbose --kubeconfig  ${HOME}/.kube/config --selector=suite=kuttlsuite --namespace="${TEST_NAMESPACE}" --service-account="${SERVICE_ACCOUNT_NAME}" --wait-time 30m ./bundle || {
       echo "****** Scorecard tests failed..."
       exit 1
    }
}

set_rbac() {
    if [ "$INSTALL_MODE" == "OwnNamespace" ]; then
        oc apply -f config/rbac/kuttl-rbac.yaml
    else 
        cp config/rbac/kuttl-rbac-cluster-wide.yaml ./
        sed -i "s/wlo-ns/${TEST_NAMESPACE}/" kuttl-rbac-cluster-wide.yaml
        oc apply -f kuttl-rbac-cluster-wide.yaml
    fi
}

unset_rbac() {
    if [ "$INSTALL_MODE" == "OwnNamespace" ]; then
        oc delete -f config/rbac/kuttl-rbac.yaml
    else 
        oc delete -f kuttl-rbac-cluster-wide.yaml
	rm kuttl-rbac-cluster-wide.yaml
    fi
}

run_kuttl_tests() {
    INSTALL_MODE=$1
    if [ "$INSTALL_MODE" == "SingleNamespace" ]; then
        set_kuttl_test_dir "kuttl-single-namespace"
    elif [ "$INSTALL_MODE" == "AllNamespaces" ]; then
	set_kuttl_test_dir "kuttl-all-namespaces"
    fi

    set_rbac
    install_operator_and_run_scorecard_tests
    result=$?
    if [[ $result != 0 ]]; then
        return $result
    fi

    if [ "$INSTALL_MODE" == "SingleNamespace" ]; then
	unset_kuttl_test_dir "kuttl-single-namespace"
    elif [ "$INSTALL_MODE" == "AllNamespaces" ]; then
        unset_kuttl_test_dir "kuttl-all-namespaces"
    fi
    unset_rbac
    uninstall_operator	
}

set_kuttl_test_dir() {
    TEST_DIR=$1
    mv bundle/tests/scorecard/kuttl bundle/tests/scorecard/kuttl-temp
    mv bundle/tests/scorecard/${TEST_DIR} bundle/tests/scorecard/kuttl
    mv bundle/tests/scorecard/kuttl-temp/kuttl-test.yaml bundle/tests/scorecard/kuttl 
}

unset_kuttl_test_dir() {
    TEST_DIR=$1
    mv bundle/tests/scorecard/kuttl/kuttl-test.yaml bundle/tests/scorecard/kuttl-temp
    mv bundle/tests/scorecard/kuttl bundle/tests/scorecard/${TEST_DIR}
    mv bundle/tests/scorecard/kuttl-temp bundle/tests/scorecard/kuttl
}

install_operator() {
    echo "****** Installing the operator in ${INSTALL_MODE} mode..."

    echo "****** Applying catalog source..."
    cat <<EOF | oc apply -f -
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: websphere-liberty-catalog
  namespace: $CONTROLLER_MANAGER_NAMESPACE
spec:
  sourceType: grpc
  image: $CATALOG_IMAGE
  displayName: WebSphere Liberty Catalog
  publisher: IBM
EOF

    if [ "$1" != "AllNamespaces" ]; then
      echo "****** Applying the OperatorGroup supporting $1..."
      cat <<EOF | oc apply -f -
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: websphere-operator-group
  namespace: $OPERATOR_GROUP_NAMESPACE
spec:
  targetNamespaces:
    - $OPERATOR_GROUP_TARGET_NAMESPACE
EOF
    else
      echo "****** Skipping OperatorGroup creation since AllNamespaces is selected..."
    fi

    echo "****** Applying the Subscription..."
    cat <<EOF | oc apply -f -
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: websphere-liberty-operator-subscription
  namespace: $CONTROLLER_MANAGER_NAMESPACE
spec:
  channel: $DEFAULT_CHANNEL
  name: ibm-websphere-liberty
  source: websphere-liberty-catalog
  sourceNamespace: $CONTROLLER_MANAGER_NAMESPACE
  installPlanApproval: Automatic
EOF
}

uninstall_operator() {
    echo "****** Uninstalling the operator..."
    oc delete subscription.operators.coreos.com websphere-liberty-operator-subscription -n ${CONTROLLER_MANAGER_NAMESPACE}

    currentCSV=$(oc get clusterserviceversion | grep ibm-websphere-liberty | cut -d ' ' -f1 )
    oc delete clusterserviceversion $currentCSV -n ${OPERATOR_GROUP_TARGET_NAMESPACE}

    oc delete catalogsource websphere-liberty-catalog -n ${CONTROLLER_MANAGER_NAMESPACE} 
    if [ "$INSTALL_MODE" != "AllNamespaces" ]; then
        oc delete operatorgroup websphere-operator-group -n ${CONTROLLER_MANAGER_NAMESPACE}
    fi
}

parse_args() {
  while [ $# -gt 0 ]; do
    case "$1" in
    -u)
      shift
      readonly DOCKER_USERNAME="${1}"
      ;;
    -p)
      shift
      readonly DOCKER_PASSWORD="${1}"
      ;;
    --cluster-url)
      shift
      readonly CLUSTER_URL="${1}"
      ;;
    --cluster-user)
      shift
      readonly CLUSTER_USER="${1}"
      ;;
    --cluster-token)
      shift
      readonly CLUSTER_TOKEN="${1}"
      ;;
    --registry-name)
      shift
      readonly REGISTRY_NAME="${1}"
      ;;
    --registry-image)
      shift
      readonly REGISTRY_IMAGE="${1}"
      ;;
    --registry-user)
      shift
      readonly REGISTRY_USER="${1}"
      ;;
    --registry-password)
      shift
      readonly REGISTRY_PASSWORD="${1}"
      ;;  
    --release)
      shift
      readonly RELEASE="${1}"
      ;;
    --test-tag)
      shift
      readonly TEST_TAG="${1}"
      ;;
    --debug-failure)
      readonly DEBUG_FAILURE=true
      ;;
    --catalog-image)
      shift
      readonly CATALOG_IMAGE="${1}"
      ;;
    --channel)
      shift
      readonly CHANNEL="${1}"
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

main "$@"
