#!/bin/bash

readonly usage="Usage: ocp-cluster-e2e.sh -u <docker-username> -p <docker-password> --cluster-url <url> --cluster-token <token> --registry-name <name> --registry-image <ns/image> --registry-user <user> --registry-password <password> --release <daily|release-tag> --test-tag <test-id> --catalog-image <catalog-image> --channel <channel>"
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
    if [[ $INSTALL_MODE = "SingleNamespace" ]]; then
      readonly INSTALL_NAMESPACE="wlo-test-single-namespace-${TEST_TAG}"
    elif [[ $INSTALL_MODE = "AllNamespaces" ]]; then
      readonly INSTALL_NAMESPACE="openshift-operators"
    else
      readonly INSTALL_NAMESPACE="wlo-test-${TEST_TAG}"
    fi

    if [ $INSTALL_MODE != "AllNamespaces" ]; then
      echo "****** Creating install namespace: ${INSTALL_NAMESPACE} for release ${RELEASE}"
      oc new-project "${INSTALL_NAMESPACE}" || oc project "${INSTALL_NAMESPACE}"
    fi

    echo "****** Creating test namespace: ${TEST_NAMESPACE} for release ${RELEASE}"
    oc new-project "${TEST_NAMESPACE}" || oc project "${TEST_NAMESPACE}"

    ## Create service account for Kuttl tests
    oc -n $TEST_NAMESPACE apply -f config/rbac/kuttl-rbac.yaml
}

## cleanup_env : Delete generated resources that are not bound to a test INSTALL_NAMESPACE.
cleanup_env() {
  ## Delete CRDs
  WLO_CRD_NAMES=$(oc get crd -o name | grep liberty.websphere | cut -d/ -f2)
  echo "*** Deleting CRDs ***"
  echo "*** ${WLO_CRD_NAMES}"
  oc delete crd $WLO_CRD_NAMES

  ## Delete Subscription
  WLO_SUBSCRIPTION_NAME=$(oc -n $INSTALL_NAMESPACE get subscription -o name | grep websphere-liberty | cut -d/ -f2)
  echo "*** Deleting Subscription ***"
  echo "*** ${WLO_SUBSCRIPTION_NAME}"
  oc -n $INSTALL_NAMESPACE delete subscription $WLO_SUBSCRIPTION_NAME

  ## Delete CSVs
  WLO_CSV_NAME=$(oc -n $INSTALL_NAMESPACE get csv -o name | grep websphere-liberty | cut -d/ -f2)
  echo "*** Deleting CSVs ***"
  echo "*** ${WLO_CSV_NAME}"
  oc -n $INSTALL_NAMESPACE delete csv $WLO_CSV_NAME

  if [ $INSTALL_MODE != "OwnNamespace" ]; then
    echo "*** Deleting project ${TEST_NAMESPACE}"
    oc delete project "${TEST_NAMESPACE}"
  fi

  if [ $INSTALL_MODE != "AllNamespaces" ]; then
    echo "*** Deleting project ${INSTALL_NAMESPACE}"
    oc delete project "${INSTALL_NAMESPACE}"
  fi
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

    if [[ -z "${INSTALL_MODE}" ]]; then
        echo "****** Missing install-mode, see usage"
        echo "${usage}"
        exit 1
    fi

    if [[ -z "${ARCHITECTURE}" ]]; then
        echo "****** Missing architecture, see usage"
        echo "${usage}"
        exit 1
    fi

    echo "****** Setting up test environment..."
    setup_env

    if [[ "${ARCHITECTURE}" != "X" ]]; then
        echo "****** Setting up tests for ${ARCHITECTURE} architecture"
        setup_tests
    fi

    if [[ -z "${DEBUG_FAILURE}" ]]; then
        trap trap_cleanup EXIT
    else
        echo "#####################################################################################"
        echo "WARNING: --debug-failure is set. If e2e tests fail, any created resources will remain"
        echo "on the cluster for debugging/troubleshooting. YOU MUST DELETE THESE RESOURCES when"
        echo "you're done, or else they will cause future tests to fail. To cleanup manually, just"
        echo "delete the namespace \"${INSTALL_NAMESPACE}\": oc delete project \"${INSTALL_NAMESPACE}\" "
        echo "#####################################################################################"
    fi

    # login to docker to avoid rate limiting during build
    echo "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME}" --password-stdin

    trap "rm -f /tmp/pull-secret-*.yaml" EXIT

    echo "****** Logging into private registry..."
    echo "${REGISTRY_PASSWORD}" | docker login ${REGISTRY_NAME} -u "${REGISTRY_USER}" --password-stdin

    echo "sleep for 3 minutes to wait for rook-cepth, knative and cert-manager to start installing, then start monitoring for completion"
    sleep 3m
    echo "monitoring knative"
    ./wait.sh deployment knative-serving
    rc_kn=$?
    echo "rc_kn=$rc_kn"
    if [[ "$rc_kn" == 0 ]]; then
        echo "knative up"
    fi

    if [[ "${ARCHITECTURE}" == "X" ]]; then
      echo "monitoring rook-ceph if architecture is ${ARCHITECTURE}"
      ./wait.sh deployment rook-ceph
      rc_rk=$?
      echo "rc_rk=$rc_rk"
      if [[ "$rc_rk" == 0 ]]; then
          echo "rook-ceph up"
      fi
    fi
    echo "****** Installing operator from catalog: ${CATALOG_IMAGE} using install mode of ${INSTALL_MODE}"
    echo "****** Install namespace is ${INSTALL_NAMESPACE}.  Test namespace is ${TEST_NAMESPACE}"    
    install_operator

    # Wait for operator deployment to be ready
    while [[ $(oc -n $INSTALL_NAMESPACE get deploy "${CONTROLLER_MANAGER_NAME}" -o jsonpath='{ .status.readyReplicas }') -ne "1" ]]; do
        echo "****** Waiting for ${CONTROLLER_MANAGER_NAME} to be ready..."
        sleep 10
    done

    echo "****** ${CONTROLLER_MANAGER_NAME} deployment is ready..."
    if [[ "$ARCHITECTURE" != "Z" ]]; then
    echo "****** Testing on ${ARCHITECTURE} so starting scorecard tests..."
    operator-sdk scorecard --verbose --kubeconfig  ${HOME}/.kube/config --selector=suite=kuttlsuite --namespace="${TEST_NAMESPACE}" --service-account="scorecard-kuttl" --wait-time 45m ./bundle || {
       echo "****** Scorecard tests failed..."
       exit 1
    }
    else
    echo "****** Testing on ${ARCHITECTURE} so running kubectl-kuttl tests..."
    kubectl-kuttl test ./bundle/tests/scorecard/kuttl --namespace "${TEST_NAMESPACE}" --timeout 200 --suppress-log=events --parallel 1 || {
       echo "****** kubectl kuttl tests failed..."
       exit 1
    }
    fi
    result=$?

    echo "****** Cleaning up test environment..."
    if [[ "${ARCHITECTURE}" != "X" ]]; then
      revert_tests
    fi
    cleanup_env

    return $result
}

install_operator() {
    # Apply the catalog
    echo "****** Applying the catalog source..."
    cat <<EOF | oc apply -f -
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: websphere-liberty-catalog
  namespace: openshift-marketplace
spec:
  sourceType: grpc
  image: $CATALOG_IMAGE
  displayName: WebSphere Liberty Catalog
  publisher: IBM
EOF

if [ $INSTALL_MODE != "AllNamespaces" ]; then
    echo "****** Applying the operator group..."
    cat <<EOF | oc apply -f -
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: websphere-operator-group
  namespace: $INSTALL_NAMESPACE
spec:
  targetNamespaces:
    - $TEST_NAMESPACE
EOF
fi

    echo "****** Applying the subscription..."
    cat <<EOF | oc apply -f -
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: websphere-liberty-operator-subscription
  namespace: $INSTALL_NAMESPACE
spec:
  channel: $DEFAULT_CHANNEL
  name: ibm-websphere-liberty
  source: websphere-liberty-catalog
  sourceNamespace: openshift-marketplace
  installPlanApproval: Automatic
EOF
}

# Substitutions to kuttl test files so that they will run on various architectures
setup_tests () {
  echo " As the architecture is ${ARCHITECTURE}..."
  if [[ "$ARCHITECTURE" == "P" ]]; then
  echo "Change affinity tests to look for ppc64le nodes"
  sed -i.bak "s,amd64,ppc64le," $(find ./bundle/tests/scorecard/kuttl/affinity -type f)
  echo "Change storage test to set storageclass to managed-nfs-storage"
  sed -i.bak "s,rook-cephfs,managed-nfs-storage," $(find ./bundle/tests/scorecard/kuttl/storage -type f)
  # These will need changing if a different image is used
  echo "Change image-stream tests to the correct digest for correct architecture"
  sed -i.bak "s,sha256:0796d9d800932a0da80d91fea720c12977bab871f8bf33b6e353b2c58aff23f1,sha256:5325d35a0c219ff545c6f906aa35b5d84a953493166c43aecd37ecc0d5e64fa6," $(find ./bundle/tests/scorecard/kuttl/image-stream -type f)
  sed -i.bak "s,sha256:5db4910bb5d5f479c55cba3ed0d9572676d50e30bf61b4a00d086f79016b8d53,sha256:f8a554c41d74dec15aab6c6f71aec741c8cb33eb2f587449e5bd7b8c46dd25b5," $(find ./bundle/tests/scorecard/kuttl/image-stream -type f)
  elif [[ "$ARCHITECTURE" == "Z" ]]; then
  echo "Change affinity tests to look for s390x nodes"
  sed -i.bak "s,amd64,s390x," $(find ./bundle/tests/scorecard/kuttl/affinity -type f)
  echo "Change storage test to set storageclass to managed-nfs-storage"
  sed -i.bak "s,rook-cephfs,managed-nfs-storage," $(find ./bundle/tests/scorecard/kuttl/storage -type f)
  # These will need changing if a different image is used
  echo "Change image-stream tests to the correct digest for correct architecture"
  sed -i.bak "s,sha256:0796d9d800932a0da80d91fea720c12977bab871f8bf33b6e353b2c58aff23f1,sha256:d622c05f4d62fc1f3cccc674c9f68cf57822c022fdd37af17bb1a7303f998ff5," $(find ./bundle/tests/scorecard/kuttl/image-stream -type f)
  sed -i.bak "s,sha256:5db4910bb5d5f479c55cba3ed0d9572676d50e30bf61b4a00d086f79016b8d53,sha256:fae90792e698d6e687c0c3b44db56f062f6374eb5fef039882308c048c9e0cbe," $(find ./bundle/tests/scorecard/kuttl/image-stream -type f)
  else
    echo "${ARCHITECTURE} is an invalid architecture type"
    exit 1
  fi
}

# As there maybe multiple runs over various architectures revert the substitutions back to amd64 defaults
revert_tests() {
  echo "Reverting test changes back to amd64"
  find ./bundle/tests/scorecard/kuttl/* -name "*.bak" -exec sh -c 'mv -f $0 ${0%.bak}' {} \;
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
    --install-mode)
      shift
      readonly INSTALL_MODE="${1}"
      ;;
        --architecture)
      shift
      readonly ARCHITECTURE="${1}"
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