#!/bin/bash
# set -xv
#
# echo "IAMTOKEN = $IAMTOKEN"

export IAMTOKEN=`cat /config/config.json | jq  ."IAMToken" | sed -e 's/"//g'`


./mkpvyaml.py

