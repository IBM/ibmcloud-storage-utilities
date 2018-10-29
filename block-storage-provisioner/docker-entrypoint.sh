#!/bin/bash
# set -xv
#
# echo "IAMTOKEN = $IAMTOKEN"

if [ -z "$IAMTOKEN" ]; then
    export IAMTOKEN=`cat /config/config.json | jq  ."IAMToken" | sed -e 's/"//g'`
fi

./mkpvyaml.py $*
