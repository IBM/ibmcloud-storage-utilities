#!/bin/bash
# shellcheck disable=SC2006,SC2002,SC2086,SC2155,SC2048
# set -xv
#
# echo "IAMTOKEN = $IAMTOKEN"

if [ -z "$IAMTOKEN" ]; then
    export IAMTOKEN=`cat /config/config.json | jq  ."IAMToken" | sed -e 's/"//g'`
fi

./mkpvyaml.py $*
