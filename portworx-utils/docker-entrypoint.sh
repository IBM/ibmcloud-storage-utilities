#!/bin/bash
# set -xv
#
# echo "IAMTOKEN = $IAMTOKEN"

export IAMTOKEN=`cat /config/config.json | jq  ."IAMToken" | sed -e 's/"//g'`


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

