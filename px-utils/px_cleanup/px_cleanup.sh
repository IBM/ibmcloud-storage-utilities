#!/bin/bash

logmessage() {
  echo "" 2>&1
  echo "$@" 2>&1
}

if ! which jq &>/dev/null
then
        echo "Jq is not installed... exiting"
        exit 1
fi

##Check ibmcloud instaled or not
if ! which ibmcloud &>/dev/null
then
        echo "IBM Cloud is not installed. Please install ibmcloud..... exiting"
        exit 1
fi

waitforportworxpodsdelete() {
DELETED=0
LIMIT=20
SLEEP_TIME=30
DESIRED=$(kubectl get ds/portworx -n kube-system -o json | jq .status.desiredNumberScheduled)
repeat=0

while [ $repeat -lt $LIMIT ] && [ $DESIRED -ne $DELETED ]; do
    DELETED=$(kubectl get pods -n kube-system -l name=portworx --field-selector status.phase=Delete -o json | jq '.items | length')
    if [ $DESIRED -eq $DELETED ]; then
        echo "(Attempt $i of $LIMIT) Portworx pods: Desired $DESIRED, Running $DELETED"
    else
        echo "(Attempt $i of $LIMIT) Portworx pods: Desired $DESIRED, Running $DELETED, sleeping $SLEEP_TIME"
        sleep $SLEEP_TIME
    fi
    repeat=$(( $repeat + 1 ))
done
echo "All the pods moved to running state"
}


rmpxservice () {

	kubectl label nodes --all px/enabled=remove --overwrite
	waitforportworxpodsdelete
	VER=$(kubectl version --short | awk -Fv '/Server Version: /{print $3}')
	kubectl delete -f "https://install.portworx.com?ctl=true&kbver=$VER"
	kubectl label nodes --all px/enabled-
}



ask() {
  # https://djm.me/ask
  local prompt default reply
    prompt="Y/n"
    default=N

  # Ask the question (not using "read -p" as it uses stderr not stdout)<Paste>
  echo -n "$1 [$prompt]:"

  # Read the answer (use /dev/tty in case stdin is redirected from somewhere else)
  read reply </dev/tty
  if [ $? -ne 0 ]; then
    logmessage "ERROR: Could not ask for user input - please run via interactive shell"
  fi

  # Default? (e.g user presses enter)
  if [ -z "$reply" ]; then
    reply=$default
  fi

  # Check if the reply is valid
  case "$reply" in
    Y*|y*) return 0 ;;
    N*|n*) return 1 ;;
    * )    echo "invalid reply: $reply"; return 1 ;;
  esac
} 
    CLUSTER_NAME=kubectl -n kube-system get cm cluster-info -o jsonpath='{.data.cluster-config\.json}' | jq -r '.name'

    if ! ask "The operation will delete Portworx components and metadata from the cluster. Do you want to continue?" N; then
          logmessage "Aborting Portworx wipe from the cluster..."
          exit 1
    else
      if ! ask "Do you want to wipeout the data also from the volumes . Please enter?" N; then
          logmessage "The operation will delete Portworx components and metadata from the cluster.The data will not be wiped out fromm the voluems..."
          rmpxservice 
      else
        if ! ask "The operation will delete Portworx components and metadata from the cluster. The operation is irreversible and will lead to DATA LOSS. Do you want to continue?" N; then
          logmessage "The operation will delete Portworx components and metadata from the cluster.The data will not be wiped out fromm the voluems..."
          rmpxservice 
       else
        logmessage "The operation will delete Portworx components and metadata and the data on the volumes..."
	bash `pwd`/scripts/px-wipe.sh -f 
       fi
      fi
   fi


echo "removing the portworx helm from the cluster"
_rc=0
helm_release=$(helm ls -A --output json | jq -r '.[]|select(.name=="portworx") | .name')
[[ -z "$helm_release" ]] && { echo "Unable to find helm release for portworx.  Ensure your helm client is at version 3 and has access to the cluster."; exit; }
helm uninstall portworx || _rc=$?
if [ $_rc -ne 0 ]; then
  logmessage "error removing the helm relese"
  exit 1;
fi
echo "Removing the Service from the catalog"
Bx_PX_svc_name=$(ibmcloud resource service-instances --service-name portworx --output json | jq -r --arg CLUSTERNAME $CLUSTER_NAME '.[]|select((.parameters.clusters==$CLUSTERNAME)) | .name')
ibmcloud resource service-instance-delete "${Bx_PX_svc_name}" -f
