#!/bin/bash

#########################################################################################
#
#
#           Build manifest list for repository/image
#           Note: Assumed to run under <operator root>/scripts
#
#
#########################################################################################

set -Eeo pipefail

readonly usage="Usage: $0 --repository <repository> --image <image> --tag <tag>"
readonly script_dir="$(dirname "$0")"

main() {
  parse_args "$@"
  check_args

  # Remove 'v' prefix from any releases matching version regex `\d+\.\d+\.\d+.*`
  if [[ "${TAG}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+ ]]; then
    readonly release_tag="${TAG#*v}"
  else
    readonly release_tag="${TAG}"
  fi
  build_manifest "${release_tag}"
}

build_manifest() {
    local tag=$1
    local target="${REGISTRY}/${IMAGE}:${tag}"

    export DOCKER_CLI_EXPERIMENTAL=enabled

    echo "Creating manifest list with $target-amd64"
    docker manifest create "$target" "$target-amd64"
    if [ "$?" != "0" ]; then
        echo "Error creating manifest list with $target-amd64"
        exit 1
    fi
    docker manifest annotate "$target" "$target-amd64" --os linux --arch amd64
    if [ "$?" != "0" ]; then
        echo "Error adding annotations for $target-amd64 to manifest list"
        exit 1
    fi

    echo "Adding $target-s390x to manifest list"
    docker manifest create --amend "$target" "$target-s390x"
    if [ "$?" != "0" ]; then
        echo "Error adding $target-s390x to manifest list"
        exit 1
    fi
    docker manifest annotate "$target" "$target-s390x" --os linux --arch s390x
    if [ "$?" != "0" ]; then
        echo "Error adding annotations for $target-s390x to manifest list"
        exit 1
    fi

    echo "Adding $target-ppc64le to manifest list"
    docker manifest create --amend "$target" "$target-ppc64le"
    if [ "$?" != "0" ]; then
        echo "Error adding $target-ppc64le to manifest list"
        exit 1
    fi
    docker manifest annotate "$target" "$target-ppc64le" --os linux --arch ppc64le
    if [ "$?" != "0" ]; then
        echo "Error adding annotations for $target-ppc64le to manifest list"
        exit 1
    fi

    docker manifest inspect "$target"
    docker manifest push "$target" --purge
    if [ "$?" = "0" ]; then
    echo "Successfully pushed $target"
    else
    echo "Error pushing $target"
    exit 1
    fi
}

check_args() {
  if [[ -z "${REGISTRY}" ]]; then
    echo "****** Missing target registry for manifest lists, see usage"
    echo "${usage}"
    exit 1
  fi

  if [[ -z "${IMAGE}" ]]; then
    echo "****** Missing target image for manifest lists, see usage"
    echo "${usage}"
    exit 1
  fi

  if [[ -z "${TAG}" ]]; then
    echo "****** Missing tag for manifest lists, see usage"
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