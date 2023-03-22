#!/bin/bash
pwd
source ./child.sh
rc=$?
echo $rc

echo $arch
echo $message

export WLO_DEMAND_ID_$arch="TEST Arch"

echo $WLO_DEMAND_ID_P
pwd


