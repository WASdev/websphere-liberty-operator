#!/usr/bin/env bash
arch=$1

if [[ "$arch" == "X" ]]; then
    wlo_demand_id=$(get_env WLO_DEMAND_ID)
    export demandId=$wlo_demand_id
    echo "calling ebc_waitForDemand.sh for X"
    
fi
if [[ "$arch" == "Z" ]]; then
    wlo_demand_id=$(get_env WLO_DEMAND_ID_Z)
    export demandId=$wlo_demand_id
    echo "calling ebc_waitForDemand.sh for Z"
fi
if [[ "$arch" == "P" ]]; then
    wlo_demand_id=$(get_env WLO_DEMAND_ID_P)
    export demandId=$wlo_demand_id
    echo "calling ebc_waitForDemand.sh for P"
fi

export ebcEnvironment=prod

json=$(./ebc_waitForDemand.sh)
rc=$?
echo "return from ebc_waitForDemand.sh for $arch"

if [[ "$rc" == 0 ]]; then
    echo "EBC create of id: $wlo_demand_id cluster successful"
else
    echo "debug of cluster may be required, issue @ebc debug $wlo_demand_id in #was-ebc channel to keep cluster for debug, issue @ebc debugcomplete $wlo_demand_id when done debugging in #was-ebc channel"
fi

status=$(jq -c '.status' <<< $json)
ip=$(jq -c '.machineAddresses.ocpinf' <<< $json)
ip=$(echo "$ip" | tr -d '"')

PRIVATE_KEY="$(get_env private_key "")"
echo -n "${PRIVATE_KEY}" | base64 -d > id_rsa

chmod 600 id_rsa
pwd
ls -l id_rsa

echo "oc version:"
oc version

token=$(ssh -o StrictHostKeyChecking=no -i id_rsa root@$ip "cat ~/auth/kubeadmin-password")

echo "json=$json"
echo "status=$status"
echo "token=$token"
echo $ip

printf '%s,%s\n' "$ip" "$token"