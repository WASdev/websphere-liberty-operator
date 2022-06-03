#!/bin/bash

# Get these values from fyre.ibm.com/account
FYRE_USER="<your_username>"
FYRE_KEY="<your_apikey>"
PRODUCT_GROUP_ID="00" # use the product group ID number, not the name

# Set these to what you want to build
CLUSTER_NAME="<your_cluster_prefix>"  # must be unique in all of Fyre
VM_OS_NAME="Ubuntu 22.04"             # see options at fyre.ibm.com/help#fyre-os
VM_SIZE="s"                           # s (2CPU, 2GB, 250GB), m (2CPU, 4GB, 250GB), l (4CPU, 8GB, 250GB), or x (8CPU, 16GB, 250GB)
VM_COUNT="1"                          # 1 if you want a single VM, but you can set up to 100

BUILD_DATA="""{
    \"type\" : \"simple\",
    \"cluster_prefix\" :\"$CLUSTER_NAME\",
    \"instance_type\" : \"virtual_server\",
    \"size\" : \"$VM_SIZE\",
    \"platform\": \"x\",
    \"os\" : \"$VM_OS_NAME\",
    \"count\" : \"$VM_COUNT\",
    \"product_group_id\": \"$PRODUCT_GROUP_ID\"
}"""

BUILD_TIME_START=$SECONDS

# Build the single-VM cluster
echo "Sending build request to Fyre..."
BUILD_REQUEST_URL="$(curl -X POST -s -k -u $FYRE_USER:$FYRE_KEY 'https://api.fyre.ibm.com/rest/v1/?operation=build' --data "$BUILD_DATA" | jq '.details' | sed 's/"//g')"

BUILD_COMPLETE_PERCENT="$(curl -s -k -u $FYRE_USER:$FYRE_KEY "https://fyre.ibm.com/embers/checkbuild?cluster=$CLUSTER_NAME")"

until [ $BUILD_COMPLETE_PERCENT -eq 100 ]
do
    BUILD_COMPLETE_PERCENT="$(curl -s -k -u $FYRE_USER:$FYRE_KEY "https://fyre.ibm.com/embers/checkbuild?cluster=$CLUSTER_NAME")"
    BUILD_REQUEST_STATUS="$(curl -s -k -u $FYRE_USER:$FYRE_KEY "https://api.fyre.ibm.com/rest/v1/?operation=query&request=showclusters" | jq --arg cluster_name "$CLUSTER_NAME" '.clusters[] | select(.name == $cluster_name) | .status' | sed 's/"//g')"
    echo "VM status is $BUILD_REQUEST_STATUS ($BUILD_COMPLETE_PERCENT%)"
    sleep 5
done

BUILD_TIME_END=$SECONDS
BUILD_TIME_DIFF=$((BUILD_TIME_END-STBUILD_TIME_STARTART))
BUILD_TIME_MINUTES=$((BUILD_TIME_DIFF/60))
BUILD_TIME_COUNTED_MINUTES=$((BUILD_TIME_MINUTES*60))
BUILD_TIME_SECONDS=$((BUILD_TIME_DIFF-BUILD_TIME_COUNTED_MINUTES))

echo "Your VM was built in $BUILD_TIME_MINUTES minutes, $BUILD_TIME_SECONDS seconds."

# Delete the cluster
curl -X POST -s -k -u $FYRE_USER:$FYRE_KEY 'https://api.fyre.ibm.com/rest/v1/?operation=delete' --data "{\"cluster_name\":\"$CLUSTER_NAME\"}" > /dev/null