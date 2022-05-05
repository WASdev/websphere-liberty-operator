#!/bin/bash

# Script to build & install the websphere liberty operator to a private registry of an OCP cluster

# -----------------------------------------------------
# Prereqs to running this script
# -----------------------------------------------------
# 1. Have "podman" or "docker" and "oc" installed & on the path
# 2. Run "oc login .." 
# 3. Run "oc registry login --skip-check"

#------------------------------------------------------------------------
# Usage
#------------------------------------------------------------------------
# dev.sh [command] [parameters]

#   Available commands:

#   all       - Run all targets
#   init      - Initialize new OCP cluster by patching registry settings and logging in
#   build     - Build and push all images
#   catalog   - Apply CatalogSource (install operator into operator hub)
#   subscribe - Apply OperatorGroup & Subscription (install operator onto cluster)

set -Eeo pipefail

readonly USAGE="Usage: dev.sh all | init | login| build | catalog | subscribe | deploy  [ -host <ocp registry hostname url> -version <operator verion to build> -image <image name> -bundle <bundle image> -catalog <catalog image> -name <operator name> -namespace <namespace> -tempdir <temp dir> ]"

main() {

  parse_args "$@"

  if [[ -z "${COMMAND}" ]]; then
    echo
    echo "${USAGE}"
    echo
    exit 1
  fi

  oc status > /dev/null 2>&1 && true
  if [[ $? -ne 0 ]]; then
    echo
    echo "Run 'oc login' to log into a cluster before running dev.sh"
    echo
    exit 1
  fi

  # Favor docker if installed. Fall back to podman. 
  # Override by setting CONTAINER_COMMAND
  docker -v > /dev/null 2>&1 && true
  if [[ $? -eq 0 ]]; then
    CONTAINER_COMMAND=${CONTAINER_COMMAND:="docker"}
    TLS_VERIFY=""
  else
    CONTAINER_COMMAND=${CONTAINER_COMMAND:="podman"}
    TLS_VERIFY=${TLS_VERIFY:="--tls-verify=false"}
  fi

  SCRIPT_DIR="$(dirname "$0")"

  # Set defaults unless overridden. 
  OCP_REGISTRY_URL=${OCP_REGISTRY_URL:=$(oc get route default-route -n openshift-image-registry --template='{{ .spec.host }}')}
  NAMESPACE=${NAMESPACE:="websphere-liberty"}
  VERSION=${VERSION:="latest"}
  if [[ "$VERSION" == "latest" ]]; then
      VVERSION=$VERSION
  else
      VVERSION=${VVERSION:=v$VERSION}
  fi
  OPERATOR_NAME=${OPERATOR_NAME:="operator"}
  IMAGE_TAG_BASE=${IMAGE_TAG_BASE:=$OCP_REGISTRY_URL/$NAMESPACE/$OPERATOR_NAME:$VVERSION}
  IMG=${IMG:=$OCP_REGISTRY_URL/$NAMESPACE/$OPERATOR_NAME:$VVERSION}
  BUNDLE_IMG=${BUNDLE_IMG:=$OCP_REGISTRY_URL/$NAMESPACE/$OPERATOR_NAME-bundle:$VVERSION}
  CATALOG_IMG=${CATALOG_IMG:=$OCP_REGISTRY_URL/$NAMESPACE/$OPERATOR_NAME-catalog:$VVERSION}
  MAKEFILE_DIR=${MAKEFILE_DIR:=$SCRIPT_DIR/..}
  TEMP_DIR=${TEMP_DIR:=/tmp}
  
  if [[ "$COMMAND" == "all" ]]; then
     init_cluster
     login_registry
     build
     bundle
     catalog
     apply_catalog
     apply_og
     apply_subscribe
  elif [[ "$COMMAND" == "init" ]]; then
     init_cluster
     login_registry
  elif [[ "$COMMAND" == "build" ]]; then
     build
     bundle
     catalog
  elif [[ "$COMMAND" == "catalog" ]]; then
     apply_catalog
  elif [[ "$COMMAND" == "subscribe" ]]; then
     apply_og
     apply_subscribe
  elif [[ "$COMMAND" == "login" ]]; then
     login_registry
  elif [[ "$COMMAND" == "deploy" ]]; then
     deploy
  else 
    echo
    echo "Command $COMMAND unrecognized."
    echo
    echo "${USAGE}"
    exit 1
  fi

}

#############################################################################
# Setup an OCP cluster to use the private registry, insecurely (testing only)
#############################################################################
init_cluster() {
    oc patch configs.imageregistry.operator.openshift.io/cluster --patch '{"spec":{"defaultRoute":true}}' --type=merge
    oc patch image.config.openshift.io/cluster  --patch '{"spec":{"registrySources":{"insecureRegistries":["'$OCP_REGISTRY_URL'"]}}}' --type=merge
    oc project $NAMESPACE > /dev/null 2>&1 && true
    if [[ $? -ne 0 ]]; then
      oc new-project $NAMESPACE 
    fi
}

login_registry() {
    echo  $CONTAINER_COMMAND login $TLS_VERIFY -u kubeadmin -p $(oc whoami -t) $OCP_REGISTRY_URL
    $CONTAINER_COMMAND login $TLS_VERIFY -u kubeadmin -p $(oc whoami -t) $OCP_REGISTRY_URL
    oc registry login --skip-check   
}

apply_catalog() {
    CATALOG_FILE=/$TEMP_DIR/catalog.yaml    
    
cat << EOF > $CATALOG_FILE
    apiVersion: operators.coreos.com/v1alpha1
    kind: CatalogSource
    metadata:
      name: websphere-liberty-catalog
      namespace: $NAMESPACE
    spec:
      sourceType: grpc
      image: $CATALOG_IMG
      imagePullPolicy: Always
      displayName: WebSphere Liberty Catalog
      publisher: IBM
      updateStrategy:
        registryPoll:
          interval: 1m
EOF

    oc apply -f $CATALOG_FILE
    rm  $CATALOG_FILE
}

apply_subscribe() {
    SUBCRIPTION_FILE=/$TEMP_DIR/subscription.yaml    
    
cat << EOF > $SUBCRIPTION_FILE
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: websphere-liberty-operator-subscription
  namespace: $NAMESPACE
spec:
  channel:  v1.0 
  name: ibm-websphere-liberty
  source: websphere-liberty-catalog
  sourceNamespace: $NAMESPACE
  installPlanApproval: Automatic
EOF

    oc apply -f $SUBCRIPTION_FILE
    rm $SUBCRIPTION_FILE          
}

apply_og() {
    OG_FILE=/$TEMP_DIR/og.yaml    
    
cat << EOF > $OG_FILE
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: websphere-operator-group
  namespace: $NAMESPACE
EOF

    oc apply -f $OG_FILE
    rm $OG_FILE          
}

###################################
# Make deploy
###################################
deploy() {
    echo "------------"
    echo "deploy"
    echo "------------"
    make -C  $MAKEFILE_DIR install VERSION=$VVERSION IMG=$IMG IMAGE_TAG_BASE=$IMAGE_TAG_BASE BUNDLE_IMG=$BUNDLE_IMG CATALOG_IMG=$CATALOG_IMG TLS_VERIFY=false
    make -C  $MAKEFILE_DIR deploy VERSION=$VVERSION IMG=$IMG IMAGE_TAG_BASE=$IMAGE_TAG_BASE BUNDLE_IMG=$BUNDLE_IMG CATALOG_IMG=$CATALOG_IMG TLS_VERIFY=false
}

###################################
# Build and push the operator image
###################################
build() {
    echo "------------"
    echo "docker-build"
    echo "------------"
    make -C  $MAKEFILE_DIR docker-build VERSION=$VVERSION IMG=$IMG IMAGE_TAG_BASE=$IMAGE_TAG_BASE BUNDLE_IMG=$BUNDLE_IMG CATALOG_IMG=$CATALOG_IMG TLS_VERIFY=false
    echo "------------"
    echo "docker-push"
    echo "------------"
    make -C  $MAKEFILE_DIR docker-push VERSION=$VVERSION IMG=$IMG IMAGE_TAG_BASE=$IMAGE_TAG_BASE BUNDLE_IMG=$BUNDLE_IMG CATALOG_IMG=$CATALOG_IMG TLS_VERIFY=false
}

###################################
# Build and push the bundle image
###################################
bundle() {
    echo "------------"
    echo "bundle"
    echo "------------"
    # Special case, Makefile make bundle only handles semantic versioning
    if [[ "$VERSION" == "latest" ]]; then
        make -C  $MAKEFILE_DIR bundle IMG=$IMG VERSION=9.9.9
        sed -i 's/$OCP_REGISTRY_URL\/$NAMESPACE\/$OPERATOR_NAME:v9.9.9/$OCP_REGISTRY_URL\/$NAMESPACE\/$OPERATOR_NAME:latest/g' $MAKEFILE_DIR/bundle/manifests/ibm-websphere-liberty.clusterserviceversion.yaml
    else
        make -C  $MAKEFILE_DIR bundle IMG=$IMG VERSION=$VERSION
    fi

    echo "------------"
    echo "bundle-build"
    echo "------------"
    make -C  $MAKEFILE_DIR bundle-build VERSION=$VVERSION IMG=$IMG IMAGE_TAG_BASE=$IMAGE_TAG_BASE BUNDLE_IMG=$BUNDLE_IMG CATALOG_IMG=$CATALOG_IMG TLS_VERIFY=false 
    echo "------------"
    echo "bundle-push"
    echo "------------"
    make -C  $MAKEFILE_DIR bundle-push VERSION=$VVERSION IMG=$IMG IMAGE_TAG_BASE=$IMAGE_TAG_BASE BUNDLE_IMG=$BUNDLE_IMG CATALOG_IMG=$CATALOG_IMG TLS_VERIFY=false
}

###################################
# Build and push the bundle image
###################################
catalog() {
    echo "------------"
    echo "catalog-build"
    echo "------------"
    make -C  $MAKEFILE_DIR catalog-build VERSION=$VVERSION IMG=$IMG IMAGE_TAG_BASE=$IMAGE_TAG_BASE BUNDLE_IMG=$BUNDLE_IMG CATALOG_IMG=$CATALOG_IMG TLS_VERIFY=false
    echo "------------"
    echo "catalog-push"
    echo "------------"
    make -C  $MAKEFILE_DIR catalog-push VERSION=$VVERSION IMG=$IMG IMAGE_TAG_BASE=$IMAGE_TAG_BASE BUNDLE_IMG=$BUNDLE_IMG CATALOG_IMG=$CATALOG_IMG TLS_VERIFY=false
}

parse_args() {
    readonly COMMAND="$1"

    while [ $# -gt 0 ]; do
    case "$1" in
    -host)
      shift
      OCP_REGISTRY_URL="${1}"
      ;;
    -namespace)
      shift
      NAMESPACE="${1}"
      ;;
    -version)
      shift
      VERSION="${1}"
      ;;
    -image)
       IMG="${1}"
      ;;
    -catalog)
       CATALOG_IMG="${1}"
      ;;
    -bundle)
       BUNDLE_IMG="${1}"
      ;;   
    -tempdir)
      shift
      TEMP_DIR="${1}"
      ;;         
    esac
    shift
  done
}

main "$@"