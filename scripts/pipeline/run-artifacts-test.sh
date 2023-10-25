#!/bin/bash
arch=$1
source scripts/pipeline/clusterWait.sh $arch
clusterurl="$ip:6443"

echo "in directory"
pwd

echo "running configure-cluster.sh"
git clone --single-branch --branch main https://$(get_env git-token)@github.ibm.com/websphere/operators.git
ls -l operators/scripts/configure-cluster/setup-ocp-cluster.sh
echo "**** issuing oc login"
oc login --insecure-skip-tls-verify $clusterurl -u kubeadmin -p $token
echo "Open Shift Console:"
console=$(oc whoami --show-console)
echo $console
echo "*** after issuing oc login"
operators/scripts/configure-cluster/setup-ocp-cluster.sh -p $token -k $(get_env ibmcloud-api-key-staging) --arch $arch -A

export PATH=$PATH:/usr/local/go/bin

# OCP test
export PIPELINE_USERNAME=$(get_env ibmcloud-api-user)
export PIPELINE_PASSWORD=$(get_env ibmcloud-api-key-staging)
export PIPELINE_REGISTRY=$(get_env pipeline-registry)
export CLUSTER_URL=$clusterurl
export CLUSTER_TOKEN=$token
export DEBUG_FAILURE=$(get_env debug-failure)

make test-artifacts-e2e
rc=$?

echo "switching back to ebc-gateway-http directory"
cd scripts/pipeline/ebc-gateway-http

if [[ "$rc" == 0 ]]; then
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
oc logout
export CLUSTER_URL=""