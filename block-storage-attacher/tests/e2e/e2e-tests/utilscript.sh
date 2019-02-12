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

[ -z "$GOPATH" ] && echo "Need GOPATH for plugin build and test executions(e.g export GOPATH=\path\to)" && exit 1


SCRIPTS_FOLDER_PATH="src/github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher/tests/e2e/scripts"
E2EPATH="src/github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher/tests/e2e/e2e-tests/"
SCRIPTS_FOLDER_PATH="$GOPATH/$SCRIPTS_FOLDER_PATH"
E2E_PATH="$GOPATH/$E2EPATH"


# Load common functions
. $SCRIPTS_FOLDER_PATH/common.sh


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
elif [ "$2" = "portworxpvcreate" ]
then
    install_portworx_plugin
    check_portworx_pod_state "portworx"
    echo "BlockVolumeAttacher-Volume-Test: PortWorx Plugin-Installation: PASS" >> $E2E_PATH/e2eTests.txt
    export CLSFILE=$3
    kubectl  create -f $CLSFILE
elif [ "$2" = "portworxdelete" ]
then
    curl -fsL https://install.portworx.com/px-wipe | bash
    helm delete --purge portworx
else
    echo "Wrong arguments"
fi




