#!/bin/bash
# shellcheck disable=SC2028,SC1091,SC2002,SC2034
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
export LC_ALL=C.UTF-8
export LANG=C.UTF-8

# Load common functions
. "$SCRIPTS_FOLDER_PATH"/common.sh

function cleanupNode ()
{
   export NodeIP=$1
   export commandstring=$2
   JOB_NAME=$(LC_CTYPE=C cat /dev/urandom | base64 | tr -dc a-z0-9 | fold -w 32 | head -n 1)
   "$E2E_PATH"/runon_worker.sh  "$NodeIP"  "$JOB_NAME" \""$commandstring"\" >>"$E2E_PATH"/command_output 2>&1
   kubectl delete job "$JOB_NAME" -n ibm-system
}



PV_Name=""
VOL_ID=""

if [ "$2" = "pvcreate" ]
 then

  export GENRATEDPVFILE=$1
  kubectl  create -f "$GENRATEDPVFILE"
elif [ "$2" = "voldelete" ]
then
    export VOL_ID=$1
    export NOD_IP=$3
    slcli block access-revoke  "$VOL_ID" --ip-address "$NOD_IP"
    slcli -y block volume-cancel  --immediate "$VOL_ID"
elif [ "$2" = "portworxpvcreate" ]
then
    kubectl apply -f "$E2E_PATH"/portworx_secret.yaml
    kubectl apply -f "$E2E_PATH"/IBM-KPbinding.yaml
    kubectl apply -f "$E2E_PATH"/IBM-KPconfig.yaml
    install_portworx_plugin
    check_portworx_pod_state "portworx"
    echo "BlockVolumeAttacher-Volume-Test: PortWorx Plugin-Installation: PASS" >> "$E2E_PATH"/e2eTests.txt
    export CLSFILE=$1
    kubectl  create -f "$CLSFILE"
elif [ "$2" = "portworxdelete" ]
then
    export NOD_IP=$3
    sudo curl -fsL https://install.portworx.com/px-wipe | bash
    kubectl delete storageclass "$1"
    helm delete --purge portworx
    kubectl  delete namespace portworx
    touch "$E2E_PATH"/command_output
    cleanupNode  "$NOD_IP"  "multipath -F"
    cleanupNode  "$NOD_IP"  "/opt/pwx/bin/pxctl sv nw --all"
    cleanupNode  "$NOD_IP"  "rm -f /etc/systemd/system/portworx*.service"
    cleanupNode  "$NOD_IP"  "grep -q '/opt/pwx/oci /opt/pwx/oci' /proc/self/mountinfo && umount /opt/pwx/oci"
    cleanupNode  "$NOD_IP"  "chattr -ie /etc/pwx/.private.json"
    cleanupNode  "$NOD_IP"  "rm -fr /opt/pwx; rm -fr /etc/pwx"
else
    echo "Wrong arguments"
fi
