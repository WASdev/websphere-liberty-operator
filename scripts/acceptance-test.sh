#!/usr/bin/env bash

set -e -o pipefail

echo "acceptance-test"

# Build e2e runner image
docker build -t e2e-runner:latest -f Dockerfile.e2e --build-arg GO_VERSION="${GO_VERSION}" . || {
	echo "Error: Failed to build e2e runner"
	exit 1
}

declare -A E2E_TESTS=(
	[ocp-e2e-run-${ARCHITECTURE}]=$(cat <<-EOF
		--volume /var/run/docker.sock:/var/run/docker.sock \
		--env PIPELINE_USERNAME=${PIPELINE_USERNAME} \
		--env PIPELINE_PASSWORD=${PIPELINE_PASSWORD} \
		--env PIPELINE_REGISTRY=${PIPELINE_REGISTRY} \
		--env PIPELINE_OPERATOR_IMAGE=${PIPELINE_OPERATOR_IMAGE} \
		--env DOCKER_USERNAME=${DOCKER_USERNAME} \
		--env DOCKER_PASSWORD=${DOCKER_PASSWORD} \
		--env CLUSTER_URL=${CLUSTER_URL} \
		--env CLUSTER_USER=${CLUSTER_USER} \
		--env CLUSTER_TOKEN=${CLUSTER_TOKEN} \
		--env TRAVIS_BUILD_NUMBER=${BUILD_NUMBER} \
		--env RELEASE_TARGET=${RELEASE_TARGET} \
		--env CATALOG_IMAGE=${CATALOG_IMAGE} \
		--env DEBUG_FAILURE=${DEBUG_FAILURE} \
		--env INSTALL_MODE=${INSTALL_MODE} \
		--env ARCHITECTURE=${ARCHITECTURE} \
		--env DIGEST=${DIGEST} \
		e2e-runner:latest \
		make test-pipeline-e2e
		EOF
	)
)

if [[ "${SKIP_KIND_E2E_TEST}" != true && "${ARCHITECTURE}" == "X" ]]; then
	E2E_TESTS[kind-e2e-run]=$(cat <<- EOF
		--volume /var/run/docker.sock:/var/run/docker.sock \
		--env FYRE_USER=${FYRE_USER} \
		--env FYRE_KEY=${FYRE_KEY} \
		--env FYRE_PASS=${FYRE_PASS} \
		--env FYRE_PRODUCT_GROUP_ID=${FYRE_PRODUCT_GROUP_ID} \
		--env TRAVIS_BUILD_NUMBER=${BUILD_NUMBER} \
		--env VM_SIZE=l \
		--env DEBUG_FAILURE=${DEBUG_FAILURE} \
		e2e-runner:latest \
		make kind-e2e-test
		EOF
	)
else
	echo "SKIP_KIND_E2E was set or architecture is not X. Skipping kind e2e..."
fi

echo "****** Starting e2e tests"
for test in "${!E2E_TESTS[@]}"; do
	docker run -d --name ${test} ${E2E_TESTS[${test}]} || {
		echo "Error: Failed to start ${test}"
		exit 1
	}
done

echo "****** Waiting for e2e tests to finish"
for test in "${!E2E_TESTS[@]}"; do
	
	# Establish monitoring variables
	monitorLoop=false
	monitorCount=1
	monitorMax=360  # Set for 360 minutes, or 6 hours  

	# wait until we are told to exit the loop either by exceeding runtime or getting an exited notice
	until [ "$monitorLoop" = true ]; do
	
		# sleep 60 seconds
        sleep 60
        
		# increment counter  
        ((monitorCount++))

        # check to see if we've exceeded time
        if  (($monitorCount>$monitorMax)); then
            monitorLoop=true
            echo "****** The max time to wait for the e2e tests to finish has elapsed"
        fi

        # check to see if tests have completed
        status="$(docker ps --all --no-trunc --filter name="^/${test}$" --format='{{.Status}}')" 
        if  ( echo "${status}" | grep -q "Exited" ); then
            monitorLoop=true
            echo "****** The e2e tests have completed"
        fi
    done

	#until docker ps --all --no-trunc --filter name="^/${test}$" --format='{{.Status}}' | grep -q Exited; do
	#	sleep 60
	#done
	
	echo "****** e2e test '${test}' have completed"
	docker logs ${test}
done

echo "****** Test results"
exit_code=0
for test in "${!E2E_TESTS[@]}"; do
	status="$(docker ps --all --no-trunc --filter name="^/${test}$" --format='{{.Status}}')"
	if echo "${status}" | grep -q "Exited (0)"; then
		echo "[PASSED] ${test}"
	else
		echo "[FAILED] ${test}: ${status}"
		exit_code=1
	fi
done
exit ${exit_code}
