#!/bin/bash

#########################################################################################
#
#
#           Script to build the multi arch images for operator
#           To skip pushing the image to the container registry, provide the `--skip-push` flag.
#           Note: Assumed to run under <operator root>/scripts
#
#
#########################################################################################

set -Eeo pipefail

readonly usage="Usage: $0 --repository <repository> --image <image> --tag <tag> [--skip-push]"

main() {
  parse_args "$@"
  check_args

  ## Define current arch variable
  case "$(uname -p)" in
  "ppc64le")
    readonly arch="ppc64le"
    ;;
  "s390x")
    readonly arch="s390x"
    ;;
  *)
    readonly arch="amd64"
    ;;
  esac

  # Remove 'v' prefix from any releases matching version regex `\d+\.\d+\.\d+.*`
  if [[ "${TAG}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+ ]]; then
    readonly release_tag="${TAG#*v}"
  else
    readonly release_tag="${TAG}"
  fi

  readonly full_image="${REGISTRY}/${IMAGE}:${release_tag}-${arch}"    

  ## build or push latest main branch
  echo "*** Building ${release_tag} for ${arch}"
  docker build -t "${full_image}" .
  if [ "$?" != "0" ]; then
      echo "Error building operator image: $full_image"
      exit 1
  fi 

  if [[ "${SKIP_PUSH}" != true ]]; then
    echo "****** Pushing image: ${full_image}"
    docker push "${full_image}"
    if [ "$?" != "0" ]; then
        echo "Error pushing operator image: $full_image"
        exit 1
    fi 
  else
    echo "****** Skipping push for operator image: $full_image"
  fi
}

check_args() {
  if [[ -z "${REGISTRY}" ]]; then
    echo "****** Missing target registry for operator build, see usage"
    echo "${usage}"
    exit 1
  fi

  if [[ -z "${IMAGE}" ]]; then
    echo "****** Missing target image for operator build, see usage"
    echo "${usage}"
    exit 1
  fi

  if [[ -z "${TAG}" ]]; then
    echo "****** Missing tag for operator build, see usage"
    echo "${usage}"
    exit 1
  fi
  
}

parse_args() {
    while [ $# -gt 0 ]; do
    case "$1" in
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