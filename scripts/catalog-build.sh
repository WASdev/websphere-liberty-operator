#!/bin/bash

OPM_TOOL="opm"
CONTAINER_TOOL="docker"

function main() {
    parse_arguments "$@"
    build_catalog
}

usage() {
    script_name=`basename ${0}`
    echo "Usage: ${script_name} [OPTIONS]"
    echo "  -n, --opm-version        [REQUIRED] Version of opm (e.g. v4.5)"
    echo "  -b, --base-image         [REQUIRED] The base image that the index will be built upon (e.g. registry.redhat.io/openshift4/ose-operator-registry)"
    echo "  -t, --output             [REQUIRED] The location where the database should be output"
    echo "  -i, --image-name         [REQUIRED] The bundle image name"
    echo "  -a, --catalog-image-name [REQUIRED] the catalog image name"
    echo "  -c, --container-tool     Tool to build image [docker, podman] (default 'docker')"
    echo "  -o, --opm-tool           Name of the opm tool (default 'opm')"
    echo "  -h, --help               Display this help and exit"
    exit 0
}


function parse_arguments() {
    if [[ "$#" == 0 ]]; then
        usage
        exit 1
    fi

    # process options
    while [[ "$1" != "" ]]; do
        case "$1" in
        -c | --container-tool)
            shift
            CONTAINER_TOOL=$1
            ;;
        -o | --opm-tool)
            shift
            OPM_TOOL=$1
            ;;    
        -n | --opm-version)
            shift
            OPM_VERSION=$1
            ;;
        -b | --base-image)
            shift
            BASE_INDEX_IMG=$1
            ;;
        -d | --directory)
            shift
            BASE_MANIFESTS_DIR=$1
            ;;
        -i | --image-name)
            shift
            BUNDLE_IMAGE=$1
            ;;
        -a | --catalog-image-name)
            shift
            CATALOG_IMAGE=$1
            ;;
        -h | --help)
            usage
            exit 1
            ;;
        -t | --output)
            shift
            TMP_DIR=$1
            ;;
        esac
        shift
    done
}

function create_empty_db() {
    mkdir -p "${TMP_DIR}/manifests"
    echo "------------ creating an empty bundles.db ---------------"
    ${CONTAINER_TOOL} run --rm -v "${TMP_DIR}":/tmp --entrypoint "/bin/initializer" "${BASE_INDEX_IMG}:${OPM_VERSION}" -m /tmp/manifests -o /tmp/bundles.db
}

function add_to_db(){
    local img=$1
    local digest="$(skopeo inspect docker://$img | grep Digest | grep -o 'sha[^\"]*')"
    local taglessImg="$(echo $img | cut -d ':' -f 1)"
    local img_digest="${taglessImg}@${digest}"
    echo "------------ adding bundle image ${img_digest} to ${TMP_DIR}/bundles.db ------------"
    "${OPM_TOOL}" registry add -b "${img_digest}" -d "${TMP_DIR}/bundles.db" -c "${CONTAINER_TOOL}" --permissive
}

function build_catalog() {
    echo "------------ Start of catalog-build ----------------"

    mkdir -p "${TMP_DIR}"
    chmod 777 "${TMP_DIR}"
    
    ##################################################################################
    ## The catalog index build will eventually require building a bundles.db file that
    ## includes all previous versions of the operator.  For now, that is not a requirement.
    ## When the time comes that another version is released and there is a need to include
    ## multiple versions of this operator, changes will be needed in this script.  See
    ## https://github.ibm.com/websphere/automation-operator/blob/main/ci/build-operator.sh 
    ## for an example on how this is done.
    ##################################################################################

    echo "Building Catalog Index Database..."

    create_empty_db
    ## for now, add the current build's image.  In future, we need to loop through all versions and add each releases bundle image
    add_to_db "${BUNDLE_IMAGE}"

    # Copy bundles.db local prior to building the image
    cp "${TMP_DIR}/bundles.db" .

    # Build catalog image 
    "${CONTAINER_TOOL}" build -t "${CATALOG_IMAGE}" -f index.Dockerfile . 
}


# --- Run ---

main $*