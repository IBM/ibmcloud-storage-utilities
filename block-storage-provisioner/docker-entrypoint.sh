#!/bin/bash
# ******************************************************************************
# * Licensed Materials - Property of IBM
# * IBM Cloud Kubernetes Service, 5737-D43
# * (C) Copyright IBM Corp. 2022 All Rights Reserved.
# * US Government Users Restricted Rights - Use, duplication or
# * disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
# ******************************************************************************

# shellcheck disable=SC2006,SC2002,SC2086,SC2155,SC2048
# set -xv
#
# echo "IAMTOKEN = $IAMTOKEN"

if [ -z "$IAMTOKEN" ]; then
    export IAMTOKEN=`cat /config/config.json | jq  ."IAMToken" | sed -e 's/"//g'`
fi

./mkpvyaml.py $*
