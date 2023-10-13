#!/bin/bash 

KEY_DIRECTORY=$1 

KEY_FILE=$2 

NAMESPACE=$3 

LTPA_SECRET_NAME=$4

ENCODING_TYPE=$5

TIME_SINCE_EPOCH_SECONDS=$(date '+%s')

PASSWORD=$(openssl rand -base64 15)

mkdir -p ${KEY_DIRECTORY}

rm -f ${KEY_FILE}

securityUtility createLTPAKeys --file=${KEY_FILE} --password=${PASSWORD} --passwordEncoding=${ENCODING_TYPE}

ENCODED_PASSWORD=$(securityUtility encode --encoding=${ENCODING_TYPE} ${PASSWORD})

APISERVER=https://kubernetes.default.svc

SERVICEACCOUNT=/var/run/secrets/kubernetes.io/serviceaccount

TOKEN=$(cat ${SERVICEACCOUNT}/token)

CACERT=${SERVICEACCOUNT}/ca.crt

SECRET_FILE=$(echo -e "{
    \"apiVersion\": \"v1\",
    \"stringData\": {
    \"lastRotation\": \"$TIME_SINCE_EPOCH_SECONDS\",
    \"password\": \"$ENCODED_PASSWORD\"
    },
    \"kind\": \"Secret\",
    \"metadata\": {
    \"name\": \"$LTPA_SECRET_NAME\",
    \"namespace\": \"$NAMESPACE\"
    },
    \"type\": \"Opaque\"
}")

curl --cacert ${CACERT} --header "Content-Type: application/json" --header "Authorization: Bearer ${TOKEN}" -X POST ${APISERVER}/api/v1/namespaces/${NAMESPACE}/secrets --data "${SECRET_FILE}"