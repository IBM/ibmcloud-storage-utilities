#!/bin/bash

# Load e2e config file
set -a
set -e
source $1
set +a

# check mandatory variables
[ -z "$GOPATH" ] && echo "Need GOPATH for plugin build and test executions(e.g export GOPATH=\path\to)" && exit 1

# Load common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"/scripts
. $SCRIPT_DIR/common.sh


# Do a bx login, user can also opt to skip this during dev-test
if [[ $TEST_BLUEMIX_LOGIN == "true" ]]; then
        echo "Bluemix Login DOne"
	bx_login
fi

# Incase of cluster_create value is "ifNotFound", then use the existing cluster (if there is one)
cluster_id=$(bx cs clusters | awk "/$PVG_CLUSTER_CRUISER/"'{print $2}')

# incase of cluster_create "always", delete the PREV cluster (if any)
if [[ -n "$cluster_id" && "$TEST_CLUSTER_CREATE" == "always" ]]; then
	# Delete the PVG_CLUSTER_CRUISER if it exists
	set +e
	rm_cluster $PVG_CLUSTER_CRUISER
	check_cluster_deleted $PVG_CLUSTER_CRUISER
	cluster_id=""
	set -e
fi

# Create cluster only if cluster is deleted/not found
if [[ -z "$cluster_id" && "$TEST_CLUSTER_CREATE" != "never" ]]; then

	# Create a cruiser
	cruiser_create $PVG_CLUSTER_CRUISER u1c.2x4 1
	
	# Put a small delay to let things settle
	sleep 30
	
	bx cs clusters
	
	# Verify cluster is up and running
	echo "Checking the cluster for deployed state..."
	check_cluster_state $PVG_CLUSTER_CRUISER
	
	echo "Checking the workers for deployed state..."
	check_worker_state $PVG_CLUSTER_CRUISER
	
	# Run sniff tests against cluster
	bx cs clusters
	bx cs cluster-get $PVG_CLUSTER_CRUISER
	bx cs workers $PVG_CLUSTER_CRUISER
	
	echo "Cluster creation is successful and ready to use"
fi

# Setup the kube configs, user can also opt to skip this during dev-test
if [[ $TEST_CLUSTER_CONFIG_DOWNLOAD == "true" ]]; then
	setKubeConfig $PVG_CLUSTER_CRUISER
	cat $KUBECONFIG
	echo "Kubeconfig file download was successful"
fi

# Update certpath from relative to full path, without which the golang test fail
addFullPathToCertsInKubeConfig
cat $KUBECONFIG
echo "Kubeclient has been configured successfully to access the cluster"


# Install helm chart (if configured). During dev-test, user might skip this, if doesn't want an override
if [[ $TEST_HELM_INSTALL == "true" ]]; then
	install_blockvolume_plugin
	check_pod_state "ibmcloud-block-storage-attacher" 
	#check_daemonset_state "ibmcloud-block-storage-driver"
fi

# Build binary (if configured), Otherwise conf must have the binary file location
if [[ $TEST_CODE_BUILD == "true" ]]; then
	cd $BLOCK_PLUGIN_HOME
        make deps
        set -oa 
        export SL_API_KEY=$PVG_SL_API_KEY
        export  SL_USERNAME=$PVG_SL_USERNAME
        bx_login
        bx cs credentials-set  --infrastructure-username  $PVG_SL_USERNAME  --infrastructure-api-key $PVG_SL_API_KEY
        bx sl init -u   $PVG_SL_USERNAME  -p  $PVG_SL_API_KEY
        bx cs init --host  $ARMADA_API_ENDPOINT
	setKubeConfig $PVG_CLUSTER_CRUISER
        export API_SERVER=$(kubectl config view | grep server | cut -f 2- -d ":" | tr -d " ")
        addFullPathToCertsInKubeConfig
	cat $KUBECONFIG
        echo "Bluemix COnfig"
        cat ~/.bluemix/config.json
	make KUBECONFIGPATH=$KUBECONFIG PVG_PHASE=$PVG_PHASE armada-portworx-e2e-test
	echo "E2E test binary was created successfully"
fi

echo "Starting ibmcloud block storage plugin e2e tests "
# Call the go binary
#$E2E_TEST_BINARY -kubeconfig $KUBECONFIG
echo "Finished ibmcloud block storage plugin e2e tests"

exit 0
