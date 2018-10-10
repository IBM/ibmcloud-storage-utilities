#!/bin/bash
# Licensed Materials - Property of IBM
#
# (C) Copyright IBM Corp. 2017 All Rights Reserved
#
# US Government Users Restricted Rights - Use, duplicate or
# disclosure restricted by GSA ADP Schedule Contract with
# IBM Corp.
# encoding: utf-8


set -x
set -e
# Skip the tests during month end
TODAY=`/bin/date +%d`
TOMORROW=`/bin/date +%d -d "1 day"`
if [ $TOMORROW -lt $TODAY ] || [ $TODAY -eq 1 ]; then
  exit 0
fi

function addFullPathToCertsInKubeConfig {
    # The e2e tests expect the full path of the certs
    # to be in the kube confg. Prior to calling this function.
    # it is expected to have a `KUBECONFIG` variable defined

    CLUSTER_DIR=$(dirname $KUBECONFIG)
    pushd ${CLUSTER_DIR}

    for certFile in $(ls | grep -E ".*.pem"); do
        certFilePATH=$(readlink -f ${certFile})
        # Replace the certs with full path unless they already have the full path
        sed -ri "s|[^\/]$certFile| $certFilePATH|g" $KUBECONFIG
    done
    popd
}


# Put a small delay to let things settle
sleep 30

mkdir -p "$GOPATH/src" "$GOPATH/bin" && sudo chmod -R 777 "$GOPATH"
mkdir -p $GOPATH/src/github.ibm.com/alchemy-containers/armada-storage-e2e
DIR="$(pwd)"
echo "Present working directory: $DIR"
cat $KUBECONFIG
ls -altr $DIR
rm -rf .git
PVG_PHASE="armada-storage"
rsync -az ./armada-storage-e2e $GOPATH/src/github.ibm.com/alchemy-containers
sed -i "s/PVG_PHASE/"$PVG_PHASE"/g" $GOPATH/src/github.ibm.com/alchemy-containers/armada-storage-e2e/common/constants.go
cd $GOPATH/src/github.ibm.com/alchemy-containers/armada-storage-e2e
ls -altr $GOPATH/src/github.ibm.com/alchemy-containers/armada-storage-e2e


addFullPathToCertsInKubeConfig
kubectl cluster-info
kubectl config view

cd $GOPATH/src/github.ibm.com/alchemy-containers/armada-storage-e2e
DIR="$(pwd)"
echo "Present working directory: $DIR" 
ls -altr $DIR

echo "Starting armada storage basic e2e tests"
export API_SERVER=$(kubectl config view | grep server | cut -f 2- -d ":" | tr -d " ")
echo $(pwd)
make KUBECONFIGPATH=$KUBECONFIG PVG_PHASE=$PVG_PHASE armada-storage-e2e-test
echo "Finished armada storage basic e2e tests"

exit 0
