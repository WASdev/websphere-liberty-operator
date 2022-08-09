#!/usr/bin/env bash

echo "Connecting to the Security Scan Toolchain"

# IMAGES_TO_SCAN is delimited by "\n" 
IMAGES_TO_SCAN=$1

echo -e "IMAGES_TO_SCAN:\n$IMAGES_TO_SCAN"

# This must be an IBM Cloud API key that has permission to run the toolchain 
IBMCLOUD_API_KEY="$(get_env ibmcloud-api-key)"

# The IBM Cloud region that is hosting the security scanning pipeline
SECSCAN_TOOLCHAIN_REGION="us-south"

# Ensure ibmcloud is updated before logging in
ibmcloud --version
ibmcloud update -f
echo "ibmcloud login --apikey $IBMCLOUD_API_KEY -r $SECSCAN_TOOLCHAIN_REGION" -a "https://cloud.ibm.com"  
ibmcloud login --apikey "$IBMCLOUD_API_KEY" -r "$SECSCAN_TOOLCHAIN_REGION" -a "https://cloud.ibm.com"    

SCANNING_PIPELINE_ID=$(get_env security-scanning-pipeline-id)

TRIGGER_NAME=$(get_env security-scanning-pipeline-trigger)
if [[ -z "${TRIGGER_NAME}" ]]; then
    TRIGGER_NAME="Security Scan Manual Trigger Multiscan"
fi

EVIDENCE_REPO=$(get_env evidence-repo)
INCIDENT_REPO=$(get_env incident-repo)
if [[ -z $EVIDENCE_REPO || -z $INCIDENT_REPO ]]; then
    TRIGGER_PROPERTIES_JSON="{\"images-to-scan\": \"$(echo ${IMAGES_TO_SCAN})\"}" 
else
    TRIGGER_PROPERTIES_JSON="{
    \"images-to-scan\": \"$(echo ${IMAGES_TO_SCAN})\", 
    \"evidence-repo\": \"${EVIDENCE_REPO}\",
    \"incident-repo\": \"${INCIDENT_REPO}\"
    }"
fi

echo "RUN_DATA=(ibmcloud dev tekton-trigger "$SCANNING_PIPELINE_ID" --trigger-name "$TRIGGER_NAME" --trigger-properties "$TRIGGER_PROPERTIES_JSON" --output json)"
RUN_DATA=$(ibmcloud dev tekton-trigger "$SCANNING_PIPELINE_ID" --trigger-name "$TRIGGER_NAME" --trigger-properties "$TRIGGER_PROPERTIES_JSON" --output json)

RUN_ID=$(echo $RUN_DATA | jq -r '.id')
echo "Security Scanning Pipeline Run ID=$RUN_ID"

MAX_TRIES=600
COMPLETE=0
for (( TRIES=0; TRIES<=$MAX_TRIES; TRIES++ ))
do
    RESULT=$(ibmcloud dev tekton-pipelinerun $SCANNING_PIPELINE_ID --run-id ${RUN_ID} --output json | jq -r '.status.state')
    if [[ $RESULT != "passed" && $RESULT != "failed" && $RESULT != "cancelled" && $RESULT != "succeeded" ]];then
        sleep 10
    else
        COMPLETE=1
        break
    fi
done
echo "Security Scanning Pipeline returned $RESULT"     
echo "Security Scanning Pipeline URL: https://cloud.ibm.com/devops/pipelines/tekton/${SCANNING_PIPELINE_ID}/runs/${RUN_ID}/build-scan-artifact/run-stage?env_id=ibm:yp:us-south"

# TODO: Add code to fail the pipeline run if the Security Scanning Pipeline returns "failed"
