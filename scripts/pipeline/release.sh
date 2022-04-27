#!/usr/bin/env bash

#
# prepare data
#

export GHE_TOKEN="$(cat ../git-token)"
export COMMIT_SHA="$(cat /config/git-commit)"
export APP_NAME="$(cat /config/app-name)"

COS_BUCKET_NAME=$(get_env COS_BUCKET_NAME "")
COS_ENDPOINT=$(get_env COS_ENDPOINT "")
export COS_URL="https://${COS_ENDPOINT}/${COS_BUCKET_NAME}/ci/${PIPELINE_RUN_ID}"
export DATE=$(date)

INVENTORY_REPO="$(cat /config/inventory-url)"
GHE_ORG=${INVENTORY_REPO%/*}
export GHE_ORG=${GHE_ORG##*/}
GHE_REPO=${INVENTORY_REPO##*/}
export GHE_REPO=${GHE_REPO%.git}

set +e
    REPOSITORY="$(cat /config/repository)"
    TAG="$(cat /config/custom-image-tag)"
set -e

export APP_REPO="$(cat /config/repository-url)"
APP_REPO_ORG=${APP_REPO%/*}
export APP_REPO_ORG=${APP_REPO_ORG##*/}

if [[ "${REPOSITORY}" ]]; then
    export APP_REPO_NAME=$(basename $REPOSITORY .git)
    APP_NAME=$APP_REPO_NAME
else
    APP_REPO_NAME=${APP_REPO##*/}
    export APP_REPO_NAME=${APP_REPO_NAME%.git}
fi

if [[ "${TAG}" ]]; then
    APP_ARTIFACTS='{ "app": "'${APP_NAME}'", "tag": "'${TAG}'", "cos": "'${COS_URL}'", "date": "'${DATE}'" }'
else
    APP_ARTIFACTS='{ "app": "'${APP_NAME}'", "cos": "'${COS_URL}'", "date": "'${DATE}'" }'
fi
#
# add to inventory
#

#cocoa inventory add \
#    --artifact="${ARTIFACT}" \
#    --repository-url="${APP_REPO}" \
#    --commit-sha="${COMMIT_SHA}" \
#    --build-number="${BUILD_NUMBER}" \
#    --pipeline-run-id="${PIPELINE_RUN_ID}" \
#    --version="$(cat /config/version)" \
#    --name="${APP_REPO_NAME}_deployment"
# loop through listed artifact images and scan each amd64 image
for artifact_image in $(list_artifacts); do
  IMAGE_ARTIFACT=$(load_artifact $artifact_image name)
  DIGEST=$(load_artifact $artifact_image digest)
  NAME="$(echo "$artifact_image" | awk '{print $1}')"

  echo "image from load_artifact:" $IMAGE_LOCATION 
  echo "arch:" $ARCH

  cocoa inventory add \
    --artifact="${IMAGE_ARTIFACT}" \
    --repository-url="${APP_REPO}" \
    --commit-sha="${COMMIT_SHA}" \
    --build-number="${BUILD_NUMBER}" \
    --pipeline-run-id="${PIPELINE_RUN_ID}" \
    --version="$(cat /config/version)" \
    --name="${NAME}" \
    --app-artifacts="${APP_ARTIFACTS}" \
    --signature="${SIGNATURE}" \
    --provenance="${IMAGE_ARTIFACT}" \
    --sha256="${DIGEST}" \
    --type="image"
done