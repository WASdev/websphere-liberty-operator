#!/bin/bash

env_setup() {
    
    echo "***** Environment setup"
    # General properties
    export GO_VERSION=$(get_env go-version)
    export PATH=$PATH:/usr/local/go/bin
    export ARCHITECTURE=$(get_env architecture)
    export TEST_ARCHITECTURE=$(get_env test-architecture)
    export OP_SHORT_NAME=$(get_env operator-short-name) 
    
    # OCP test properties
    export PIPELINE_USERNAME=$(get_env ibmcloud-api-user)
    export PIPELINE_PASSWORD=$(get_env ibmcloud-api-key-staging)
    export PIPELINE_REGISTRY=$(get_env pipeline-registry)
    export DEBUG_FAILURE=$(get_env debug-failure)

    export PATH=$PATH:/usr/local/go/bin

    # Install go if needed
    make setup-go GO_RELEASE_VERSION=$GO_VERSION

    # Determin the architectures to be tested
    if [[ "$TEST_ARCHITECTURE" == "ZXP" && "$ARCHITECTURE" == "ZXP" ]]; then
        architecture_types=(X P Z)
    else
        architecture_types=(X)
    fi

    BASE_DIR=$(pwd)
    echo "BASE_DIR: ${BASE_DIR}"
}

setup_cluster() {

    # For each architecture wait for the ocp cluser to be provisioned and then set up the cluser
    echo "***** Cluster setup"
    for types in "${architecture_types[@]}"; do
        echo "Running clusterWait.sh to wait for ${types} cluster to be provisioned"
        pwd
        source ${BASE_DIR}/scripts/pipeline/clusterWait.sh ${types}
        clusterurl="$ip:6443"
        export CLUSTER_URL_${types}=$clusterurl
        export CLUSTER_TOKEN_${types}=$token
        echo "**** issuing oc login"
        oc login --insecure-skip-tls-verify $clusterurl -u kubeadmin -p $token
        console=$(oc whoami --show-console)
        export CLUSTER_CONSOLE_${types}=$console
        echo "Open Shift Console url is: $console"
        echo "CLUSTER_TOKEN_${types} is: $token"
        echo "CLUSTER_URL_${types} is: $clusterurl"
        echo "*** after issuing oc login"
        RELEASE_ACCEPTANCE_TEST=$(get_env release-acceptance-test)
        if [[ ! -z "$RELEASE_ACCEPTANCE_TEST" && "$RELEASE_ACCEPTANCE_TEST" != "false" && "$RELEASE_ACCEPTANCE_TEST" != "no"  ]]; then
        CLUSTER_CONFIG_OPTIONS=" --skip-create-icsp"
        fi
        echo "Running setup-ocp-cluster.sh"
        ${BASE_DIR}/operators/scripts/configure-cluster/setup-ocp-cluster.sh -p $token -k $(get_env ibmcloud-api-key-staging) --arch $arch -A $CLUSTER_CONFIG_OPTIONS
    done
}


setup_tests() {
    echo "***** Test setup"
    cd ${BASE_DIR}

    # Creating the test directories for kuttl
    mkdir -p bundle/tests/scorecard/kuttl
    mkdir -p bundle/tests/scorecard/kind-kuttl

    # Copying all the relevent kuttl test and config file
    cp operators/tests/config.yaml bundle/tests/scorecard/
    cp -rf operators/tests/common/* bundle/tests/scorecard/kuttl
    if [[ ${OP_SHORT_NAME} == "wlo" ]]; then
        cp -rf operators/tests/all-liberty/* bundle/tests/scorecard/kuttl
        cp -rf operators/tests/websphere-liberty/* bundle/tests/scorecard/kuttl
    elif [[ ${OP_SHORT_NAME} == "olo" ]]; then
        cp -rf operators/tests/all-liberty/* bundle/tests/scorecard/kuttl
    fi

    # Copying all the kind only kuttl tests. 
    cp -rf operators/tests/kind/* bundle/tests/scorecard/kind-kuttl

    docker build -t e2e-runner:latest -f Dockerfile.e2e --build-arg GO_VERSION="${GO_VERSION}" . || {
	echo "Error: Failed to build e2e runner"
	exit 1
    }
    declare -Axg E2E_TESTS
    for types in "${architecture_types[@]}"
    do
        url=CLUSTER_URL_${types}
        token=CLUSTER_TOKEN_${types}
        E2E_TESTS[ocp-e2e-run-${types}]=$(cat <<- EOF
--volume /var/run/docker.sock:/var/run/docker.sock \
--env PIPELINE_USERNAME=$(get_env ibmcloud-api-user) \
--env PIPELINE_PASSWORD=$(get_env ibmcloud-api-key-staging) \
--env PIPELINE_REGISTRY=$(get_env pipeline-registry) \
--env CLUSTER_URL=${!url} \
--env CLUSTER_TOKEN=${!token} \
--env OP_SHORT_NAME=$(get_env operator-short-name) \
--env DEBUG_FAILURE=$(get_env debug-failure) \
--env ARCHITECTURE=${types} \
e2e-runner:latest \
make test-artifacts-e2e
EOF
        )        
    E2E=E2E_TESTS[ocp-e2e-run-${types}]
    done
}

run_tests() {

    echo "***** Starting e2e tests"
    for test in "${!E2E_TESTS[@]}"; do
        echo "Starting: $test"
        docker run -d --name ${test} ${E2E_TESTS[${test}]} || {
            echo "Error: Failed to start ${test}"
            exit 1
        }
    done
    echo "***** Waiting for e2e tests to finish"
    	
    # Establish monitoring variables
    monitorLoop=false
    monitorCount=1
    monitorMax=240  # Set for 240 minutes, or 4 hours  
    declare -Axg testCompleted
    # wait until we are told to exit the loop either by exceeding runtime or getting an exited notice
    until [ "$monitorLoop" = true ]; do
    
        # sleep 60 seconds
        sleep 60
        
        # increment counter  
        ((monitorCount++))

        # check to see if we've exceeded time
        if  (($monitorCount>$monitorMax)); then
            monitorLoop=true
            echo "***** The max time to wait for the e2e tests to finish has elapsed"
        fi
        # check to see if tests have completed
        for test in "${!E2E_TESTS[@]}"; do
            status="$(docker ps --all --no-trunc --filter name="^/${test}$" --format='{{.Status}}')" 
            if ( echo "${status}" | grep -q "Exited (0)" ); then
                if [ "${testCompleted[${test}]}" != "PASSED" ]; then
                    testCompleted[${test}]="PASSED"
                    echo -e "\033[0;32m***** e2e test '${test}' has completed successfully\033[0m"
                fi
            elif ( echo "${status}" | grep -q "Exited" ); then
                if [ -z "${testCompleted[${test}]}" ]; then
                    testCompleted[${test}]=$status
                    echo -e "\033[1;31m***** e2e test '${test}' has completed with errors\033[0m"
                fi
            fi
        done
        if [[ ${#testCompleted[@]} -eq ${#E2E_TESTS[@]} ]]; then
            monitorLoop=true
            echo "***** All tests completed"
        fi

    done

    echo "***** Test results"
    for test in "${!E2E_TESTS[@]}"; do
        echo -e "\033[0;32m***** Start of ${test} logs\033[0m"
        docker logs ${test}
        sleep 30
        echo "***** End of ${test} logs"
    done
    echo "***** Overall test results"
    for test in "${!E2E_TESTS[@]}"; do
        if [[ ${testCompleted[${test}]} = "PASSED" ]]; then
            echo -e "\033[0;32m[PASSED] ${test}\033[0m"
        else
            echo -e "\033[1;31m[FAILED] ${test}: ${testCompleted[${test}]}\033[0m"
        fi
    done
}

cleanup() {
    echo "***** Cleanup"
    for types in "${architecture_types[@]}"; do
        demandId=$(get_env $(echo "$operator" | tr '[:lower:]' '[:upper:]')_DEMAND_ID_${types})
        if [[ ${testCompleted[ocp-e2e-run-${types}]} = "PASSED" ]]; then
            echo "Acceptance tests passed so deleting cluster with demandId: ${demandId}"
            ${BASE_DIR}/ebc-gateway-http/ebc_complete.sh
        else
            hours=$(get_env ebc_autocomplete_hours "6")
            console=CLUSTER_CONSOLE_${types}
            token=CLUSTER_TOKEN_${types}
            echo "Your acceptance test failed, the cluster will be retained for $hours hours."
            echo "debug of cluster may be required, issue @ebc debug ${demandId} in #was-ebc channel to keep cluster for debug"
            echo "issue @ebc debugcomplete ${demandId} when done debugging in #was-ebc channel "
            echo "access console at: ${!console}"
            echo "credentials: kubeadmin/${!token}"
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
            curl -X POST -H 'Content-type: application/json' --data '{"text":"issue @ebc debug '${demandId}' in #was-ebc channel to keep cluster for debug"}' $(get_env slack_web_hook_url) </dev/null
            curl -X POST -H 'Content-type: application/json' --data '{"text":"access console at: '${!console}'"}' $(get_env slack_web_hook_url) </dev/null
            curl -X POST -H 'Content-type: application/json' --data '{"text":"credentials: kubeadmin/'${!token}'"}' $(get_env slack_web_hook_url) </dev/null
        fi 
    done
}

main() {
    echo "*** Artifact acceptance test stage ***"

    echo "Setting up test environment"
    env_setup
    setup_cluster
    setup_tests

    run_tests
    cleanup
}

main $*
