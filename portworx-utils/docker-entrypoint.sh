#!/bin/bash
# set -xv
#
# echo "IAMTOKEN = $IAMTOKEN"

if [ -z $IAMTOKEN ]; then
    export IAMTOKEN=`cat /config/config.json | jq  ."IAMToken" | sed -e 's/"//g'`
fi


if [ $# -eq 0 ]
then
   ./mkpvyaml.py
elif [ "$1" = "vls" ]
then
   ./vls
elif [ "$1" = "dvol" ]
then
    shift
    ./dvol $*
fi

