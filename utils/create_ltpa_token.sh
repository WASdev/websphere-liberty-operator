#!/bin/bash 

keyDirectory=$1 

keyFile=$2 

NAME=$3

NAMESPACE=$4 

LTPA_SECRET_NAME=$5

LAST_ROTATION=abc

PASSWORD=test

mkdir -p ${keyDirectory}

rm -f ${keyFile}

securityUtility createLTPAKeys --file=${keyFile} --password=${PASSWORD} --passwordEncoding=aes

ENCODED_PASSWORD=$(securityUtility encode --encoding=aes ${PASSWORD})

APISERVER=https://kubernetes.default.svc

SERVICEACCOUNT=/var/run/secrets/kubernetes.io/serviceaccount

TOKEN=$(cat ${SERVICEACCOUNT}/token)

CACERT=${SERVICEACCOUNT}/ca.crt

SECRET_FILE=$(echo -e "{
    \"apiVersion\": \"v1\",
    \"stringData\": {
    \"lastRotation\": \"$LAST_ROTATION\",
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