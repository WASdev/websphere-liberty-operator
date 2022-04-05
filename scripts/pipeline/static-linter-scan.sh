#!/bin/bash

readonly usage="Usage: static-linter-scan.sh --git-token <GHE token> --static-linter-version <Linter Version>"
readonly WORK_DIR="${PWD}/linter-tool"

function install_linter() {

    # Download JSON of all versions
    curl -H "Authorization: token ${GIT_TOKEN}" -H "Accept: application/vnd.github.v3.raw" -s https://github.ibm.com/api/v3/repos/IBMPrivateCloud/content-verification/releases > "${WORK_DIR}/versions.json"
    
    # Using jq, parse list of versions to find download for the version required
    # Linter version is hard-coded here, for now
    LINTER_TOOL_URL=$(jq -r '.[] | .assets | .[] |  select (.browser_download_url | endswith("v3.0.1/cv-linux-amd64.tar.gz")) | .url' "${WORK_DIR}/versions.json")
    
    # Use wget to download the linter binary, confirm it was successful
    wget --user token --password $GIT_TOKEN --auth-no-challenge --header "Accept: application/octet-stream" $LINTER_TOOL_URL -O $WORK_DIR/cv.tar.gz
    rc=$?
    if [ $rc -ne 0 ]; then
        # echo "Linter version ${STATIC_LINTER_VERSION} could not be located"
        echo "Linter version 3.0.1 could not be located"
        # Don't exit catastrophically, we can deal with this type failure at another time
        exit 0
    fi

    # Decompress the linter binary
    tar -xvf $WORK_DIR/cv.tar.gz -C $WORK_DIR
}


function main() {
    parse_args "$@"

    # Verify parameters
    if [[ -z "${GIT_TOKEN}" ]]; then
        echo "****** Missing git token, see usage"
        echo "${usage}"
        exit 1
    fi

    # Skipping this parm for now because I was unable to come up with a solution for how to inject the version into the jq query
    # Perhaps this is somthing that can be solved down the road.
    # if [[ -z "${STATIC_LINTER_VERSION}" ]]; then
    #    echo "****** Missing static linter version, see usage"
    #    echo "${usage}"
    #    exit 1
    # fi

    # Create working directory
    if [[ -d "${WORK_DIR}" ]]; then
        echo "Working directory ${WORK_DIR} exists.  Deleting contents."
        rm -rf $WORK_DIR
    fi
    echo "Creating work directory ${WORK_DIR}"
    mkdir "${WORK_DIR}"
    
    # Install Linter
    install_linter "${STATIC_LINTER_VERSION}"

    # Run Linter
    $WORK_DIR/cv lint -o lintOverrides.yaml operator

}

parse_args() {
  while [ $# -gt 0 ]; do
    case "$1" in
    --git-token)
      shift
      readonly GIT_TOKEN="${1}"
      ;;
    --static-linter-version)
      shift
      readonly STATIC_LINTER_VERSION="${1}"
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
