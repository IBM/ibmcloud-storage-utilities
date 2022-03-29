#!/bin/bash
# shellcheck disable=SC2162,SC2181,SC2086,SC2236,SC2162,SC2068,SC2006,SC2001,SC2034,SC2207,SC2184,SC2069

# Adds portworx to a cluster using etcd as the kvdb database.
#
#
# Note: Currently only existing etcd services are supported.  If you don't have an etcd service
#       instance, create one before running this tool.


shopt -s expand_aliases
alias ic="ibmcloud"

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


CLUSTER=$1
CLUSTER_CHECK=$(kubectl -n kube-system get cm cluster-info -o jsonpath='{.data.cluster-config\.json}' | jq -r '.name')
echo "${CLUSTER_CHECK}"
[[ -z "$CLUSTER_CHECK" ]] && { echo "Unable to determine cluster name, Either the cluser does not exist or kube config is not set."; exit; }


echo "Gathering information for cluster ${CLUSTER} ..."
CLUSTER_ID=$(ic cs cluster get --cluster ${CLUSTER} --json | jq -r '.id')


APIKEY_NAME="pxtmpapikey"



px_svc_name=$(ic resource service-instances --service-name portworx --output json | jq -r --arg CLUSTERNAME $CLUSTER '.[]|select((.parameters.clusters==$CLUSTERNAME)) | .name')
[[ ! -z "$px_svc_name" ]] && { echo "Error: Portworx service '$px_svc_name' is already installed on cluster $CLUSTER. Setup your kube config to a cluster which does not have portworx." ; exit; }

read -p "About to install portworx on cluster $CLUSTER. Enter y to continue: " continue
[[ ! "y" = "$continue" ]] && exit


echo "Gathering information for portworx service..."
px_region=$(ic cs cluster get --cluster ${CLUSTER} --json | jq -r '.region')
px_name="Portworx-Enterprise-${CLUSTER}"
px_user=$(ic target --output json | jq -r '.user.display_name')
owner=$(ic target --output json | jq -r '.account.owner' | cut -d@ -f1)
api_key=$(ic iam  api-key-create ${APIKEY_NAME}  --output json | jq -r '.apikey')
px_cluster_name="${CLUSTER}-pw"

IFS=$'\n' etcd=($(ic resource service-instances --service-name databases-for-etcd -q))
# Remove the column headers
unset etcd[0]
echo ""
echo "Select an etcd database instance to use for your portworx cluster:"
echo ""
if [[ ${#etcd[@]} -gt 1 ]]; then
  index="0"
  for key in ${!etcd[@]}; do
    ((index++))
    printf '%s. %s\n' $index "${etcd[key]}"
  done
    ((index++))
  printf '%s. %s\n\n' "99" "Exit"
  read -p "Enter etcd service instance number or new to create a new etcd instance> " selected_etcd
  [[ "99" = "${selected_etcd}" ]] && exit
  i=0
  ((i=selected_etcd))
  etcd_svc=${etcd[$i]}
else
  etcd_svc=${etcd[1]}
fi

etcd_svc_name=${etcd_svc%% *}

echo "Gathering etcd information for $etcd_svc_name..."
crn=$(ic resource service-instance "$etcd_svc_name" --output json | jq -r '.[0] | .crn')
conn=$(ic resource service-keys --output json | jq -r --arg crn "$crn" '.[] | select(.source_crn==$crn) | .credentials.connection.grpc.composed[0]')
ep="etcd:https://${conn#*@}"
uid_pwd=${conn%@*}
pwd=${uid_pwd##*:}
uid_tmp=${uid_pwd%:*}
uid=${uid_tmp#*//}
cert=$(ic resource service-keys --output json | jq -r --arg crn "$crn" '.[] | select(.source_crn==$crn) | .credentials.connection.grpc.certificate.certificate_base64')
uid=$(printf "%s" "$uid" | base64)
pwd=$(printf "%s" "$pwd" | base64)
pwd_join_string=`echo $pwd | sed 's/\\n//g'`
IFS=' ' read -ra  string_array <<< "$pwd_join_string"
pwd="${string_array[0]}${string_array[1]}"



echo "Creating etcd secret in the kube-system namespace..."
(cat << EOF
---
apiVersion: v1
kind: Secret
metadata:
  name: px-etcd-certs
  namespace: kube-system
type: Opaque
data:
  ca.pem: ${cert}
  username: ${uid}
  password: ${pwd}
EOF
) | if ! kubectl create -f - 2>&1 > /dev/null; then
    echo "unable to create Secrets.. exiting ....."
    exit 1
fi

target_group=$(ic target --output json | jq -r '.resource_group.name // empty')
[[ -z "${target_group}" ]] && target_group="default"

echo "Creating the portworx service $px_name in $px_region using resource group $target_group for $px_user"

parms=$(jq -n --arg cl "$CLUSTER" \
              --arg apikey "$api_key" \
              --arg pxcluster "$px_cluster_name" \
              --arg etcd_ep $ep \
              --arg portworx_version "2.5.7" \
              '{apikey: $apikey,cluster_name: $pxcluster,clusters: $cl,internal_kvdb: "external",etcd_endpoint: $etcd_ep,etcd_secret: "px-etcd-certs",secret_type: "k8s",portworx_version: "2.6.2.1"}')

ic resource service-instance-create "${px_name}" portworx px-enterprise $px_region -g $target_group --parameters "${parms}"
