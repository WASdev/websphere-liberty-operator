#!/bin/bash

GH_BRANCH=ciorch-integration
GH_REPOSITORY=build-liberty-images-ubi
GH_ORG=WEBSTEC1
CI_TRIGGER=wldocker
CI_CONFIG_FILE=.ci-orchestrator/websphere-liberty-build.yml
pipelineName=Websphere Liberty Docker Container Build


function main() {
    parse_arguments "$@"
    request_ciorchestrator
}

function print_usage() {
    script_name=`basename ${0}`
    echo "Usage: ${script_name} [OPTIONS]"
    echo ""
    echo "Kick off of CI Orchestrator job"
    echo ""
    echo "Options:"
    echo "   -u, --user       string  IntranetId to use to authenticate to CI Orchestrator"
    echo "   --password       string  Intranet Password to use to authenticate to CI Orchestrator"
    echo "   -b, --branch     string  Github Repository branch"
    echo "   -r, --repository string  GitHub Repository to use"
    echo "   --org            string  Github Organisation containing repository"
    echo "   --trigger        string  Name of trigger within CI Orchestrator config file"
    echo "   --configFile     string  Location of CI Orchestrator config file"
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
        -u | --user)
            shift
            USER=$1
            ;;
        --password)
            shift
            PASSWORD=$1
            ;;
        -b | --branch)
            shift
            GH_BRANCH=$1
            ;;
        -r | --repository)
            shift
            GH_REPOSITORY=$1
            ;;
        --org)
            shift
            GH_ORG=$1
            ;;
        --trigger)
            shift
            CI_TRIGGER=$1
            ;;
        --configFile)
            shift
            CI_CONFIG_FILE=$1
            ;;                        
        -h | --help)
            print_usage
            exit 1
            ;;
        esac
        shift
    done
}

function request_ciorchestrator() {
    pipelineId=OnePipeline_${PIPELINE_RUN_ID}_${RANDOM}
    cat >ciorchestrator-submit.json <<EOL
    {
        "type": "PipelineTriggered",
        "ecosystemRouting": "dev",
        "pipelineId": "${pipelineId}",
        "pipelineName": "${pipelineName}",
        "triggerName": "${CI_TRIGGER}",
        "triggerType": "manual",
        "requestor": "${USER}",
        "properties": {
            "scriptBranch": "${GH_BRANCH}",
            "scriptOrg": "${GH_ORG}"
        },
        "configMetadata": {
            "apiRoot": "https://github.ibm.com/api/v3",
            "org": "${GH_ORG}",
            "repo": "${GH_REPOSITORY}",
            "branch": "${GH_BRANCH}",
            "filePath": "${CI_CONFIG_FILE}"
        }
    }
    EOL

    echo "${pipelineId}" >ciorchestrator-submit.id

    echo "Sending Pipeline Request to CI Orchestrator pipelineId: ${pipelineId} as ${USER}"
    curl -v -X POST \
        -H "Content-Type: application/json"  \
        -d @ciorchestrator-submit.json \
        -u "${USER}:${PASSWORD}"
        https://libh-proxy1.fyre.ibm.com/eventPublish/rawCIData/${pipelineId}

}


# --- Run ---

main $*