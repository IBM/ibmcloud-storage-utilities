#!/bin/bash
# Licensed Materials - Property of IBM
#
# (C) Copyright IBM Corp. 2017 All Rights Reserved
#
# US Government Users Restricted Rights - Use, duplicate or
# disclosure restricted by GSA ADP Schedule Contract with
# IBM Corp.
# encoding: utf-8

#export  GENRATEDPVFILE="/root/Port_Worx_E2e/src/github.ibm.com/alchemy-containers/armada-storage-e2e/e2e-tests/pv-muraliportworx.yaml"


PV_Name=""
VOL_ID=""

if [ "$2" = "pvcreate" ]
 then

  export GENRATEDPVFILE=$1
  kubectl  create -f $GENRATEDPVFILE
elif [ "$2" = "voldelete" ]
then
    export VOL_ID=$1
    export NOD_IP=$3
    bx sl block access-revoke -p $NOD_IP $VOL_ID
    bx sl  block volume-cancel -f --immediate $VOL_ID
else
    echo "Wrong arguments"
fi




