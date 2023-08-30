#!/bin/bash
arch=$1
source ./clusterWait.sh $arch
clusterurl="$ip:6443"

cd ../..
echo "in directory"
pwd

## git clone --single-branch --branch cosolidate-tests https://$(get_env git-token)@github.ibm.com/websphere/operators.git
ls -l operators/scripts/configure-cluster/configure-cluster.sh
echo "**** issuing oc login"
oc login --insecure-skip-tls-verify $clusterurl -u kubeadmin -p $token
echo "Open Shift Console:"
console=$(oc whoami --show-console)
echo $console
echo "*** after issuing oc login" 
RELEASE_ACCEPTANCE_TEST=$(get_env release-acceptance-test)
if [[ ! -z "$RELEASE_ACCEPTANCE_TEST" && "$RELEASE_ACCEPTANCE_TEST" != "false" && "$RELEASE_ACCEPTANCE_TEST" != "no"  ]]; then
  CLUSTER_CONFIG_OPTIONS=" --skip-create-icsp"
fi
echo "running configure-cluster.sh"
operators/scripts/configure-cluster/configure-cluster.sh -p $token -k $(get_env ibmcloud-api-key-staging) --arch $arch -A $CLUSTER_CONFIG_OPTIONS

export GO_VERSION=$(get_env go-version)
make setup-go GO_RELEASE_VERSION=$GO_VERSION
export PATH=$PATH:/usr/local/go/bin
export INSTALL_MODE=$(get_env install-mode)
export ARCHITECTURE=$arch

# OCP test
export PIPELINE_USERNAME=$(get_env ibmcloud-api-user)
export PIPELINE_PASSWORD=$(get_env ibmcloud-api-key-staging)
export PIPELINE_REGISTRY=$(get_env pipeline-registry)
export PIPELINE_OPERATOR_IMAGE=$(get_env pipeline-operator-image)
export DOCKER_USERNAME=$(get_env docker-username)
export DOCKER_PASSWORD=$(get_env docker-password)
#export CLUSTER_URL=$(get_env test-cluster-url)
export CLUSTER_URL=$clusterurl
#export CLUSTER_USER=$(get_env test-cluster-user kubeadmin)
export CLUSTER_TOKEN=$token
export RELEASE_TARGET=$(get_env branch)
export DEBUG_FAILURE=$(get_env debug-failure)

# Kind test
export SKIP_KIND_E2E_TEST=$(get_env SKIP_KIND_E2E_TEST)
export FYRE_USER=$(get_env fyre-user)
export FYRE_KEY=$(get_env fyre-key)
export FYRE_PASS=$(get_env fyre-pass)
export FYRE_PRODUCT_GROUP_ID=$(get_env fyre-product-group-id)

# acceptance-test.sh return values
export KIND_E2E_TEST=1
export OCP_E2E_X_TEST=2
export OCP_E2E_P_TEST=4
export OCP_E2E_Z_TEST=8
export UNKNOWN_E2E_TEST=256

echo "${PIPELINE_PASSWORD}" | docker login "${PIPELINE_REGISTRY}" -u "${PIPELINE_USERNAME}" --password-stdin
if [[ ! -z "$RELEASE_ACCEPTANCE_TEST" && "$RELEASE_ACCEPTANCE_TEST" != "false" && "$RELEASE_ACCEPTANCE_TEST" != "no"  ]]; then
  RELEASE_TARGET=$(curl --silent "https://api.github.com/repos/WASdev/websphere-liberty-operator/releases/latest" | jq -r .tag_name)
  PIPELINE_PRODUCTION_IMAGE=$(get_env pipeline-production-image)
  IMAGE="${PIPELINE_PRODUCTION_IMAGE}:${RELEASE_TARGET}"
  export CATALOG_IMAGE="${PIPELINE_PRODUCTION_IMAGE}-catalog:${RELEASE_TARGET}"
else
  IMAGE="${PIPELINE_REGISTRY}/${PIPELINE_OPERATOR_IMAGE}:${RELEASE_TARGET}"
  export CATALOG_IMAGE="${PIPELINE_REGISTRY}/${PIPELINE_OPERATOR_IMAGE}-catalog:${RELEASE_TARGET}"
fi
echo "one-pipeline Image value: ${IMAGE}"
echo "one-pipeline Catalog Image value: ${CATALOG_IMAGE}"
DIGEST="$(skopeo inspect docker://$IMAGE | grep Digest | grep -o 'sha[^\"]*')"

export DIGEST
echo "one-pipeline Digest Value: ${DIGEST}"

echo "setting up tests from operators repo - runTest.sh"
mkdir -p bundle/tests/scorecard/kuttl
mkdir -p bundle/tests/scorecard/kind-kuttl

# Copying all the relevent kuttl test and config file
cp operators/tests/config.yaml bundle/tests/scorecard/
cp -rf operators/tests/common/* bundle/tests/scorecard/kuttl
cp -rf operators/tests/all-liberty/* bundle/tests/scorecard/kuttl
cp -rf operators/tests/websphere-liberty/* bundle/tests/scorecard/kuttl

# Copying all the kind only kuttl tests. Deciding if the run is a kind run is done in the acceptance-test.sh script
cp -rf operators/tests/kind/* bundle/tests/scorecard/kind-kuttl

# Copying the common test scripts
mkdir scripts/test
cp -rf operators/scripts/test/* scripts/test

cd ../..
echo "directory before acceptance-test.sh"
pwd
echo "Getting the operator short name"
export OP_SHORT_NAME=$(get_env operator-short-name)
echo "Operator shortname is: ${OP_SHORT_NAME}"
echo "Running modify-tests.sh script"
scripts/test/modify-tests.sh --operator ${OP_SHORT_NAME} --arch ${ARCHITECTURE}

echo "Running acceptance-test.sh script"
scripts/acceptance-test.sh
rc=$?
keep_cluster=0

if (( (rc & OCP_E2E_X_TEST) >0 )) || 
   (( (rc & OCP_E2E_P_TEST) >0 )) ||
   (( (rc & OCP_E2E_Z_TEST) >0 )) ||
   (( (rc & UNKNOWN_E2E_TEST) >0 )); then
   keep_cluster=1
fi

echo "switching back to ebc-gateway-http directory"
cd scripts/pipeline/ebc-gateway-http

if [[ "$keep_cluster" == 0 ]]; then
    if (( (rc & KIND_E2E_TEST) >0 )); then
      slack_users=$(get_env slack_users)
      echo "slack_users=$slack_users"
      eval "arr=($slack_users)"
      for user in "${arr[@]}"; do 
        echo "user=$user"
        curl -X POST -H 'Content-type: application/json' --data '{"text":"<'$user'>  kind accceptance test failure see below "}' $(get_env slack_web_hook_url)
        echo " "
      done
      pipeline_url="https://cloud.ibm.com/devops/pipelines/tekton/${PIPELINE_ID}/runs/${PIPELINE_RUN_ID}?env_id=ibm:yp:us-south"
      curl -X POST -H 'Content-type: application/json' --data '{"text":"Your kind acceptance test failed."}' $(get_env slack_web_hook_url) </dev/null
      curl -X POST -H 'Content-type: application/json' --data '{"text":"Failing pipeline: '$pipeline_url'"}' $(get_env slack_web_hook_url) </dev/null
    fi
    ./ebc_complete.sh
else
    hours=$(get_env ebc_autocomplete_hours "6")
    echo "Your acceptance test failed, the cluster will be retained for $hours hours."
    echo "debug of cluster may be required, issue @ebc debug $wlo_demand_id in #was-ebc channel to keep cluster for debug"
    echo "issue @ebc debugcomplete $wlo_demand_id when done debugging in #was-ebc channel "
    echo "access console at: $console"
    echo "credentials: kubeadmin/$token"
    slack_users=$(get_env slack_users)
    echo "slack_users=$slack_users"
    eval "arr=($slack_users)"
    for user in "${arr[@]}"; do 
      echo "user=$user"
      curl -X POST -H 'Content-type: application/json' --data '{"text":"<'$user'>  accceptance test failure see below "}' $(get_env slack_web_hook_url)
      echo " "
    done
    pipeline_url="https://cloud.ibm.com/devops/pipelines/tekton/${PIPELINE_ID}/runs/${PIPELINE_RUN_ID}?env_id=ibm:yp:us-south"
    curl -X POST -H 'Content-type: application/json' --data '{"text":"Your acceptance test failed."}' $(get_env slack_web_hook_url) </dev/null
    curl -X POST -H 'Content-type: application/json' --data '{"text":"Failing pipeline: '$pipeline_url'"}' $(get_env slack_web_hook_url) </dev/null
    curl -X POST -H 'Content-type: application/json' --data '{"text":"The cluster will be retained for '$hours' hours.  If you need more time to debug ( 72 hours ):"}' $(get_env slack_web_hook_url) </dev/null
    curl -X POST -H 'Content-type: application/json' --data '{"text":"issue @ebc debug '$wlo_demand_id' in #was-ebc channel to keep cluster for debug"}' $(get_env slack_web_hook_url) </dev/null
    curl -X POST -H 'Content-type: application/json' --data '{"text":"access console at: '$console'"}' $(get_env slack_web_hook_url) </dev/null
    curl -X POST -H 'Content-type: application/json' --data '{"text":"credentials: kubeadmin/'$token'"}' $(get_env slack_web_hook_url) </dev/null
fi

echo "Cleaning up after tests have be completed"
echo "switching back to scripts/pipeline directory"
cd ..
echo "Deleting test scripts ready for another run..."
rm -rf ../../bundle/tests/scorecard/kuttl/*
rm -rf ../../bundle/tests/scorecard/kind-kuttl/*
oc logout
export CLUSTER_URL=""