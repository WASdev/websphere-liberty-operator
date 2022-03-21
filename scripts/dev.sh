#!/bin/bash

# Script to build & install the websphere liberty operator to a private registry of an OCP cluster

# Prereqs to running this script
# -----------------------------------------------------
# 1. Have "podman" and "oc" installed & on the PATH
# 2. Run "oc login .."" 
# 3. Run "oc registry login --skip-check"

set -Eeo pipefail

# dev.sh  init = initialize new OCP cluster
# dev.sh  build = build and push all images
# dev.sh  all = Run all targets
# dev.sh  catalog = Apply CatalogSource (install operator into operator hub)
# dev.sh  subscribe = Apply Subsciption (install operator)

readonly USAGE="Usage: dev.sh all | init | build | catalog | subscribe  [ -host <ocp registry hostname url> -version <operator verion to build> -image <image name> -bundle <bundle image> -catalog <catalog imager> -name <operator name> -namespace <namespace> ]"

main() {

  parse_args "$@"

  if [[ -z "${COMMAND}" ]]; then
    echo
    echo "${USAGE}"
    echo
    exit 1
  fi
  
  if [[ -z "${OCP_REGISTRY_URL}" ]]; then
    echo
    echo "OCP registry hostname is required. Set CLUSTER_REGISTRY_URL or use the -host command line option."
    echo
    echo "${USAGE}"
    exit 1
  fi

  # Set defaults unless overridden. 
  NAMESPACE=${NAMESPACE:="websphere-liberty"}
  VERSION=${VERSION:="0.0.1"}
  VVERSION=${VVERSION:=v$VERSION}
  OPERATOR_NAME=${OPERATOR_NAME:="operator"}
  IMAGE_TAG_BASE=${IMAGE_TAG_BASE:=$OCP_REGISTRY_URL/$NAMESPACE/$OPERATOR_NAME:$VVERSION}
  IMG=${IMG:=$OCP_REGISTRY_URL/$NAMESPACE/$OPERATOR_NAME:$VVERSION}
  BUNDLE_IMG=${BUNDLE_IMG:=$OCP_REGISTRY_URL/$NAMESPACE/$OPERATOR_NAME-bundle:$VVERSION}
  CATALOG_IMG=${CATALOG_IMG:=$OCP_REGISTRY_URL/$NAMESPACE/$OPERATOR_NAME-catalog:$VVERSION}
  MAKEFILE_DIR=${MAKEFILE_DIR:=..}
  
  if [[ "$COMMAND" == "all" ]]; then
     init_cluster
     login_registry
     build
     apply_catalog
     apply_subscribe
  elif [[ "$COMMAND" == "init" ]]; then
     init_cluster
  elif [[ "$COMMAND" == "build" ]]; then
     build
     bundle
     catalog
  elif [[ "$COMMAND" == "catalog" ]]; then
     apply_catalog
  elif [[ "$COMMAND" == "subscribe" ]]; then
     apply_subscribe
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
    oc patch image.config.openshift.io/cluster  --patch '{"spec":{"registrySources":{"insecureRegistries":["$OCP_REGISTRY_URL"]}}}' --type=merge
    oc patch configs.imageregistry.operator.openshift.io/cluster --patch '{"spec":{"defaultRoute":true}}' --type=merge
    oc new-project $NAMESPACE
}

login_registry() {
    HOST=$(oc get route default-route -n openshift-image-registry --template='{{ .spec.host }}')
    podman login -u kubeadmin -p $(oc whoami -t) --tls-verify=false $HOST  
    oc registry login --skip-check   
}

apply_catalog() {
    CATALOG_FILE=/tmp/catalog.yaml    
    
cat << EOF > $CATALOG_FILE
    apiVersion: operators.coreos.com/v1alpha1
    kind: CatalogSource
    metadata:
      name: websphere-liberty-catalog
      namespace: $NAMESPACE
    spec:
      sourceType: grpc
      image: $CATALOG_IMG
      displayName: WebSphere Liberty Catalog
      publisher: IBM
      updateStrategy:
        registryPoll:
          interval: 1m
EOF

    oc apply -f $CATALOG_FILE
    rm  $CATALOG_FILE
}

apply_operator() {
    SUBCRIPTION_FILE=/tmp/subscription.yaml    
    
cat << EOF > $SUBCRIPTION_FILE
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: websphere-liberty-operator-subscription
  namespace: $NAMESPACE
spec:
  channel:  v1 
  name: websphere-liberty
  source: websphere-liberty-catalog
  sourceNamespace: $NAMESPACE
  installPlanApproval: Automatic
EOF

    oc apply -f $SUBCRIPTION_FILE
    rm $SUBCRIPTION_FILE          
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
    make -C  $MAKEFILE_DIR bundle IMG=$IMG VERSION=$VERSION
    echo "------------"
    echo "build-build"
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
    esac
    shift
  done
}

main "$@"