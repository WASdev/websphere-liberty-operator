#!/bin/bash
# This script is intended to be used by CI Tools like Jenkins to interact with the HTTP Gateway
# This script requests an environment asynchronously
# Parameters will be supplied via environment variables
# Example Call
# -------------
# export intranetId_USR=myIntranet
# export intranetId_PSW=myPassword
# export demandId=A Uniquely Generated ID
# export ebcEnvironment=prod (or dev to use dev ebc)
# export ebc_plan=hur-plainJenkins-ubuntu18_x86
# export ebc_priority=90
# export ebc_autoCompleteAfterXHours=1
# export ebc_reasonForEnvironment=MyTestBuild
# export ebc_jenkins_agent_label=MyLabel
# export ebc_jenkins_instance_name=MyJenkinsInstance
# ./ebc_demand.sh
# -------------
# Any ebc_ environment will be considered part of the demand

ebcEnvironment="${ebcEnvironment:-prod}"
if [ $ebcEnvironment = prod ] ; then
    gateway_url="${gateway_url_override:-https://libh-proxy1.fyre.ibm.com/ebc-gateway-http}"
else
    gateway_url="${gateway_url_override:-https://libh-proxy1.fyre.ibm.com/ebc-gateway-http-$ebcEnvironment}"
fi
demandJson="{"
for var in "${!ebc_@}"; do
        demandJson="${demandJson}\"$var\":\"${!var}\","
done

demandJson="${demandJson}\"ebc_createdBy\":\"ebc_demand.sh\"}"
echo "Creating $demandId via ${gateway_url} using $demandJson"
curl -v --fail -s -X POST -d "${demandJson}" -H "Content-type: application/json" -u "$intranetId_USR:$intranetId_PSW" --insecure ${gateway_url}/environments/${demandId}
rc=$?
if [[ $rc -ne 0 ]]; then
  echo "Issue sending EBC Request giving up.  Curl returned $rc"
  exit 1;
fi
pending=true


# Wait for a few minutes to see if it is impossible, otherwise just let CI system queue
count=0
while $pending;  do
  status=$(curl -v --fail -s -X GET -s -u "$intranetId_USR:$intranetId_PSW" --insecure ${gateway_url}/environments/${demandId}/status)
  if [[ $status != "PENDING" && $status != "" ]]; then
    pending=false;
  elif [[ $count -gt 60 ]]; then
    #Bail after 60 seconds
    pending=false
  else
    sleep 10
    count=$((count+10))
  fi
done


if [[ $status == "LIVE" ]]; then
	echo "Spinup Success final result was:${status}"
elif [[ $status == "PENDING" ]]; then
	echo "Spinup in progress"    
else 
    echo "Spinup Failed final result was:${status}"
    exit 1;
fi