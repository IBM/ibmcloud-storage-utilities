#!/bin/bash

TALISMAN_IMAGE=portworx/talisman
TALISMAN_TAG=latest
WIPE_CLUSTER="--wipecluster"
MAX_RETRIES=60
TIME_BEFORE_RETRY=5 #seconds
JOB_NAME=talisman
KUBECTL_EXTRA_OPTS=""
WIPER_IMAGE=portworx/px-node-wiper
WIPER_TAG=latest


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
          WIPE_CLUSTER=""
      else
        if ! ask "The operation will delete Portworx components and metadata from the cluster. The operation is irreversible and will lead to DATA LOSS. Do you want to continue?" N; then
          logmessage "The operation will delete Portworx components and metadata from the cluster.The data will not be wiped out fromm the voluems..."
          WIPE_CLUSTER=""
       else
        logmessage "The operation will delete Portworx components and metadata and the data on the volumes..."
        WIPE_CLUSTER="--wipecluster"
       fi
      fi
   fi

##########################
# Added IS_OPENSHIFT variable to check if cluster is an openshift cluster
##########################
if [[ $CLUSTER_VERSION == *"openshift"* ]]; then
  command -v oc
  if [ $? -eq 0 ]; then
    _out=$(command -v oc 2>&1 && oc version 2>&1)
    if [ $? -eq 0 ]; then
      echo "Detected OpenShift system. Adding talisman-account user to privileged scc"
      oc adm policy add-scc-to-user privileged system:serviceaccount:kube-system:talisman-account
      if [ $? -ne 0 ]; then
        logmessage "failed to add talisman-account to privileged scc. exit code: $?"
      fi
    else
      echo "oc binary found but oc version command failed. Not using OpenShift mode."
    fi
  fi
fi 
##########################

VER=$(kubectl version --short | awk -Fv '/Server Version: /{print $3}')
if [ -z "$VER" ]; then
	logmessage "failed to get kubernetes server version. Make sure you have kubectl setup on current machine."
fi


VER=( ${VER//./ } )
echo "Parsed version is "${VER[0]}.${VER[1]}""

if [ "${VER[0]}.${VER[1]}" == "1.7" ] || [ "${VER[0]}.${VER[1]}" == "1.6" ]; then
  logmessage "This script doesn't support wiping Portworx from Kubernetes $VER clusters."
fi

CLIENT_VER=$(kubectl version --short | awk -Fv '/Client Version: /{print $2}')
if [ -z "$CLIENT_VER" ]; then
  logmessage "failed to get kubernetes client version. Make sure you have kubectl setup on current machine."
fi

CLIENT_VERI=$(echo $CLIENT_VER | awk -F. '{print $1*100+$2}')
echo "Parsed client version (int) is ${CLIENT_VERI}"
if [ $CLIENT_VERI -lt 114 ]; then
  KUBECTL_EXTRA_OPTS="--show-all"
else
  KUBECTL_EXTRA_OPTS=""
fi

kubectl delete -n kube-system job talisman 2>/dev/null || true

RETRY_CNT=0
while true; do
  PODS=$(kubectl get pods -n kube-system -l name=$JOB_NAME $KUBECTL_EXTRA_OPTS 2>/dev/null)
  if [ $? -eq 0 ]; then
    NUM_PODS=$(echo -n "$PODS" | grep -c -v NAME)
    if [ $NUM_PODS -eq 0 ]; then
      break
    fi
  fi

  RETRY_CNT=$((RETRY_CNT+1))
  if [ $RETRY_CNT -ge $MAX_RETRIES ]; then
    logmessage "failed to delete old talisman pods"
  fi

  sleep $TIME_BEFORE_RETRY
done


cat <<EOF | kubectl apply -f -
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: talisman-account
  namespace: kube-system
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: talisman-role-binding
subjects:
- kind: ServiceAccount
  name: talisman-account
  namespace: kube-system
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io
---

apiVersion: batch/v1
kind: Job
metadata:
  name: $JOB_NAME
  namespace: kube-system
spec:
  backoffLimit: 1
  template:
    metadata:
      labels:
        name: $JOB_NAME
    spec:
      serviceAccount: talisman-account
      containers:
      - name: $JOB_NAME
        image: $TALISMAN_IMAGE:$TALISMAN_TAG
        args: ["-operation",
               "delete",
               "$WIPE_CLUSTER",
               "-wiperimage",
               "$WIPER_IMAGE",
               "-wipertag",
               "$WIPER_TAG"]
        imagePullPolicy: Always
        volumeMounts:
        - name: etcpwx
          mountPath: /etc/pwx
      volumes:
      - name: etcpwx
        hostPath:
          path: /etc/pwx
      restartPolicy: Never
EOF

echo "Talisman job for wiping Portworx started. Monitor logs using: 'kubectl logs -n kube-system -l job-name=talisman'"

NUM_DESIRED=1
RETRY_CNT=0
while true; do
  NUM_SUCCEEDED=0
  NUM_FAILED=0
  PODS=$(kubectl get pods -n kube-system -l name=$JOB_NAME 2>/dev/null)
  if [ $? -eq 0 ]; then
    CREATING=$(echo "$PODS" | grep ContainerCreating)
    if [ ! -z "$CREATING" ]; then
      echo "Pod that will perform wipe of Portworx is still in container creating phase"
    else
      NUM_FAILED=$(kubectl get job -n kube-system talisman $KUBECTL_EXTRA_OPTS -o jsonpath='{.status.failed}' 2>/dev/null)
      if [ $? -eq 0 ]; then
        if [ ! -z "$NUM_FAILED" ] && [ $NUM_FAILED -ge 1 ]; then
          kubectl logs -n kube-system -l name=$JOB_NAME
          logmessage "Job to wipe px cluster failed."
        fi
      fi

      NUM_SUCCEEDED=$(kubectl get job -n kube-system talisman $KUBECTL_EXTRA_OPTS -o jsonpath='{.status.succeeded}' 2>/dev/null)
      if [ ! -z "$NUM_SUCCEEDED" ] && [ $NUM_SUCCEEDED -eq $NUM_DESIRED ]; then
        break
      fi

      echo "waiting on $JOB_NAME to complete..."
      RUNNING_POD=$(echo "$PODS" | grep Running | awk '/^talisman/{print $1}')
      if [ ! -z "$RUNNING_POD" ]; then
        echo "Monitoring logs of pod: $RUNNING_POD"
        kubectl logs -n kube-system -f $RUNNING_POD
      fi
    fi
  fi

  RETRY_CNT=$((RETRY_CNT+1))
  if [ $RETRY_CNT -ge $MAX_RETRIES ]; then
    kubectl logs -n kube-system -l name=$JOB_NAME
    logmessage "Timed out trying to wipe Portworx cluster."
  fi

  sleep $TIME_BEFORE_RETRY
done

echo "Portworx cluster wipe succesfully completed."
_rc=0
kubectl delete job -n kube-system talisman                    || _rc=$?
kubectl delete serviceaccount -n kube-system talisman-account || _rc=$?
kubectl delete clusterrolebinding talisman-role-binding       || _rc=$?
kubectl delete crd volumeplacementstrategies.portworx.io
if [ $_rc -ne 0 ]; then
   logmessage "error cleaning up pods"
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
