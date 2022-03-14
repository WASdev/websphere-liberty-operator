#!/bin/bash

OPM_TOOL="opm"
CONTAINER_TOOL="docker"

function main() {
    parse_arguments "$@"
    build_catalog
}

function print_usage() {
    script_name=`basename ${0}`
    echo "Usage: ${script_name} [OPTIONS]"
    echo ""
    echo "Build catalog index"
    echo ""
    echo "Options:"
    echo "   -t, --token      string  Travis API token"
    echo "   -b, --branch     string  Github Repository branch"
    echo "   -l, --launch             Launch Travis job"
    echo "   -m, --monitor            Monitor Travis job"
    echo "   -n, --buildnum    string Build Identification"
    echo "   -r, --repository string  GitHub Repository to use"
    echo "   -c, --commit     string  GH head commit ID"
    echo "   -h, --help               Print usage information"
    echo ""
}


function parse_arguments() {
    if [[ "$#" == 0 ]]; then
        print_usage
        exit 1
    fi

    # process options
    while [[ "$1" != "" ]]; do
        case "$1" in
        -t | --token)
            shift
            TRAVIS_TOKEN=$1
            ;;
        -b | --branch)
            shift
            BRANCH=$1
            ;;    
        -l | --launch)
            LAUNCH_TRAVIS=yes
            ;;
        -m | --monitor)
            MONITOR_TRAVIS=yes
            ;;
        -r | --repository)
            shift
            GH_REPO=$1
            ;;
        -n | --buildnum)
            shift
            BUILD_NUMBER=$1
            ;;
        -c | --commit)
            shift
            GH_COMMIT_ID=$1
            ;;
        -h | --help)
            print_usage
            exit 1
            ;;
        esac
        shift
    done
}

function create_empty_db() {
    mkdir -p "${TMP_DIR}/manifests"
    echo "creating a empty bundles.db ..."
    ${CONTAINER_TOOL} run --rm -v "${TMP_DIR}":/tmp --entrypoint "/bin/initializer" "${BASE_INDEX_IMG}:${OPM_VERSION}" -m /tmp/manifests -o /tmp/bundles.db
}

function add_to_db(){
    local img=$1
    echo "------------ adding bundle image ${img} to db ------------"
    "${OPM_TOOL}" registry add -b "${img}" -d "${TMP_DIR}/bundles.db" "${OPM_DEBUG_FLAG}" -c "${CONTAINER_TOOL}"
}


function build_catalog() {
    echo "*********** Start of catalog-build ************"


    #mkdir -p "${TMP_DIR}"
    #chmod 777 "${TMP_DIR}"
    
    # configure podman with redirects
    sudo cp ./scripts/registries.conf /etc/containers/registries.conf

    #################################################
    ## The following section will be needed once 
    ## this operator has more than one release.  The
    ## script referenced can be found here:
    ## https://github.ibm.com/Justin-Fleming/operator-build-scripts/blob/1b00db2bdc1b34506824dba80ced6a3c40e6f993/versionsAndBundlesFromReleases.sh
    ## For the sake of completing this build scripting
    ## in time for a release, bypassing this port for
    ## now.  This should be ported eventually.
    #################################################
    #echo "Creating versions.txt file..."
    #mkdir -p ~/tmpdir
    #export TMPDIR=~/tmpdir
    #./submodules/operator-build-scripts/versionsAndBundlesFromReleases.sh -y bundle/manifests/ibm-websphere-automation.clusterserviceversion.yaml -a bundle/metadata/annotations.yaml --no-bundle-images -t "${TAG_PATTERN}"

    # Change 'version' value in versions.txt to match DOCKER_TAG. If this is a GA release build (ie. CSV_VERSION == DOCKER_TAG) no change will occur.
    #OLD_LINE=$(grep "\"$CSV_VERSION\"" versions.txt)
    #NEW_LINE=$(echo $OLD_LINE | sed "s/\"$CSV_VERSION\"/\"$DOCKER_TAG\"/")
    #sed -i.bak "s/\"$CSV_VERSION\".*/$NEW_LINE/" versions.txt

    #echo "versions.txt:"
    #cat versions.txt

    # Populate .env file with CP_CREDS value (Note, if we support multiple architectures in the future, we would need to update .env with Z_NODE and/or P_NODE here as well)
    #CP_CREDS="${W3_USERNAME,,}:${ARTIFACTORY_API_KEY}"
    #sed -i "s/CP_CREDS=.*/CP_CREDS=\"${CP_CREDS}\"/" ./submodules/operator-build-scripts/.env

    echo "Building Catalog Index Database..."
    #sudo ./submodules/operator-build-scripts/buildIndexDatabase.sh icr.io/cpopen -n v4.9 -b registry.redhat.io/openshift4/ose-operator-registry --versions-csv ./versions.txt --image-name websphere-automation-operator-bundle -t $TMPDIR/operator-build --image-tag-prefix "" --container-tool podman --debug
    #cp $TMPDIR/operator-build/bundles.db .

    # build catalog image
    #$HOME/build.common/bin/docker-build.sh -n $NAMESPACE -i $CATALOG_IMAGE_NAME -t $DOCKER_TAG -b . -f index.Dockerfile

    # publish catalog image
    #$BASE_DIR/docker-push.sh -r $DOCKER_REPO -n $NAMESPACE -i $CATALOG_IMAGE_NAME -t $DOCKER_TAG
}


# --- Run ---

main $*