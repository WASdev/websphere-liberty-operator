#!/bin/bash

# =================================================================================================
# Login to Docker Repo
# =================================================================================================
echo "${PIPELINE_PASSWORD}" | docker login "${PIPELINE_REGISTRY}" -u "${PIPELINE_USERNAME}" --password-stdin
echo "${REDHAT_PASSWORD}" | docker login "${REDHAT_REGISTRY}" -u "${REDHAT_USERNAME}" --password-stdin

if [[ "$DISABLE_ARTIFACTORY" == "false" ]]; then
    echo "${ARTIFACTORY_TOKEN}" | docker login "${ARTIFACTORY_REPO_URL}" -u "${ARTIFACTORY_USERNAME}" --password-stdin
fi

# =================================================================================================
# Install / configure operator specific dependencies
# =================================================================================================
# Install operator-sdk
make setup