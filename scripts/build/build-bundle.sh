#!/bin/bash

#########################################################################################
#
#
#           Script to bundle the multi arch images for operator
#           To skip pushing the image to the container registry, provide the `--skip-push` flag.
#           Note: Assumed to run under <operator root>/scripts
#
#
#########################################################################################

set -Eeo pipefail

readonly usage="Usage: $0 --repository <repository> --image <image> --prod-image <prod-repository/image> --tag <tag> [--skip-push]"

main() {
  parse_args "$@"
  check_args

  # Remove 'v' prefix from any releases matching version regex `\d+\.\d+\.\d+.*`
  if [[ "${TAG}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+ ]]; then
    readonly release_tag="${TAG#*v}"
  else
    readonly release_tag="${TAG}"
  fi

  readonly digest="$(skopeo inspect docker://${REGISTRY}/${IMAGE}:${release_tag} | grep Digest | grep -o 'sha[^\"]*')"
  readonly full_image="${PROD_IMAGE}@${digest}"
  readonly bundle_image="${REGISTRY}/${IMAGE}-bundle:${release_tag}" 
  
  echo "*** Bundling ${full_image}.  Bundle location will be ${bundle_image}."
  make bundle bundle-build IMG="${full_image}" BUNDLE_IMG="${bundle_image}" 
  if [ "$?" != "0" ]; then
        echo "Error building bundle image: $bundle_image"
        exit 1
  fi 

  if [[ "${SKIP_PUSH}" != true ]]; then
    echo "****** Pushing bundle: ${bundle_image}"
    make bundle-push IMG="${full_image}" BUNDLE_IMG="${bundle_image}"
    if [ "$?" != "0" ]; then
        echo "Error pushing bundle image: $bundle_image"
        exit 1
    fi 
  else
    echo "****** Skipping push for bundle image: ${bundle_image}"
  fi
}

check_args() {  
  if [[ -z "${REGISTRY}" ]]; then
    echo "****** Missing target registry for bundle build, see usage"
    echo "${usage}"
    exit 1
  fi

  if [[ -z "${IMAGE}" ]]; then
    echo "****** Missing target image for bundle build, see usage"
    echo "${usage}"
    exit 1
  fi

  if [[ -z "${PROD_IMAGE}" ]]; then
    echo "****** Missing production image reference for bundle, see usage"
    echo "${usage}"
    exit 1
  fi

  if [[ -z "${TAG}" ]]; then
    echo "****** Missing tag for bundle build, see usage"
    echo "${usage}"
    exit 1
  fi
}

parse_args() {
    while [ $# -gt 0 ]; do
    case "$1" in
    --prod-image)
      shift
      readonly PROD_IMAGE="${1}"
      ;;
    --registry)
      shift
      readonly REGISTRY="${1}"
      ;;
    --image)
      shift
      readonly IMAGE="${1}"
      ;;
    --skip-push)
      readonly SKIP_PUSH=true
      ;;
    --tag)
      shift
      readonly TAG="${1}"
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