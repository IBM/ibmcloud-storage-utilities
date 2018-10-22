#!/bin/bash
set -xv
#
# CLUSTER=`grep cluster: /data/yamlgen.yaml | awk '{print $2}'`
# mkdir /etc/kubeadmin
# cp /config/plugins/container-service/clusters/$CLUSTER/* /etc/kubeadmin 
# export KUBECONFIG=`ls /etc/kubeadmin/*.yml`
# echo KUBECONFIG=$KUBECONFIG
# kubectl get nodes  || { echo "Can't access Kubernetes through KUBECONFIG=$KUBECONFIG" >&2 ; exit 1; }
export IAMTOKEN=`cat /config/config.json | jq  ."IAMToken" | sed -e 's/"//g'`
echo "IAMTOKEN = $IAMTOKEN"


./mkpvyaml

