#!/bin/bash
# Load e2e config file
set -a
set -xe
source $1
set +a

# check mandatory variables
[ -z "$GOPATH" ] && echo "Need GOPATH for plugin build and test executions(e.g export GOPATH=\path\to)" && exit 1

SCRIPTS_FOLDER_PATH="src/github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher/scripts/"
E2EPATH="src/github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher/tests/e2e/e2e-tests/"
SCRIPTS_FOLDER_PATH="$GOPATH/$SCRIPTS_FOLDER_PATH"
E2E_PATH="$GOPATH/$E2EPATH"
MKPVYAML="mkpvyaml"
YAMLPATH="yamlgen.yaml"
MKPVYAML="$SCRIPTS_FOLDER_PATH$MKPVYAML"
YAMLPATH="$SCRIPTS_FOLDER_PATH$YAMLPATH"

if [ "$PVG_CLUSTER_TYPE" == "classic" ]; then

    if [ `echo $PVG_CLUSTER_KUBE_VERSION | grep -c  "openshift" ` -gt 0 ]; then
       CLUSTERTYPE="ROKS Classic"
    else
       CLUSTERTYPE="IKS Classic"
    fi
elif [ "$PVG_CLUSTER_TYPE" == "vpc-classic" ]; then

    if [ `echo $PVG_CLUSTER_KUBE_VERSION | grep -c  "openshift" ` -gt 0 ]; then
       CLUSTERTYPE="ROKS VPC Gen1"
    else
       CLUSTERTYPE="IKS VPC Gen1"
    fi
else
    if [ `echo $PVG_CLUSTER_KUBE_VERSION | grep -c  "openshift" ` -gt 0 ]; then
       CLUSTERTYPE="ROKS VPC Gen2"
    else
       CLUSTERTYPE="IKS VPC Gen2"
    fi
fi



CLUSTERDETAILS=" Cluster Type:$CLUSTERTYPE \n Region:$ARMADA_REGION \n Cluster Location:$PVG_CLUSTER_LOCATION \n Kube-Version:$PVG_CLUSTER_KUBE_VERSION \n"
echo -e "$CLUSTERDETAILS" > $E2E_PATH/setupDetails.txt

# Load common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"/scripts
. $SCRIPT_DIR/common.sh


# Do a bx login, user can also opt to skip this during dev-test
if [[ $TEST_BLUEMIX_LOGIN == "true" ]]; then
        echo "Bluemix Login DOne"
	bx_login
        if [ $ARMADA_REGION == "us-south" ]; then
          ibmcloud ks  api $IMAGE_REGISTRY
        fi
        ibmcloud cr login
fi

# Incase of cluster_create value is "ifNotFound", then use the existing cluster (if there is one)
cluster_id=$(ibmcloud ks clusters | awk "/$PVG_CLUSTER_CRUISER/"'{print $2}')

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
	
        ibmcloud ks clusters	
	
	# Verify cluster is up and running
	echo "Checking the cluster for deployed state..."
	check_cluster_state $PVG_CLUSTER_CRUISER
	
	echo "Checking the workers for deployed state..."
	check_worker_state $PVG_CLUSTER_CRUISER
	
	# Run sniff tests against cluster
	ibmcloud ks clusters
	ibmcloud ks cluster get --cluster $PVG_CLUSTER_CRUISER
	ibmcloud ks workers --cluster $PVG_CLUSTER_CRUISER
	
	echo "Cluster creation is successful and ready to use"
fi

        echo "BlockVolumeAttacher-Volume-Test: Cluster-Creation: PASS" > $E2E_PATH/e2eTests.txt

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
	check_pod_state "ibm-block-storage-attacher" 
	#check_daemonset_state "ibmcloud-block-storage-driver"
fi
       

echo "BlockVolumeAttacher-Volume-Test: Plugin-Installation: PASS" >> $E2E_PATH/e2eTests.txt

# Build binary (if configured), Otherwise conf must have the binary file location
if [[ $TEST_CODE_BUILD == "true" ]]; then
	cd $BLOCK_PLUGIN_HOME
        make deps
        set -oa 
        export SL_API_KEY=$PVG_SL_API_KEY
        export  SL_USERNAME=$PVG_SL_USERNAME
        bx_login
        #bx cs credentials-set  --infrastructure-username  $PVG_SL_USERNAME  --infrastructure-api-key $PVG_SL_API_KEY
        #bx sl init -u   $PVG_SL_USERNAME  -p  $PVG_SL_API_KEY
        #bx cs init --host  $ARMADA_API_ENDPOINT
	setKubeConfig $PVG_CLUSTER_CRUISER
        export API_SERVER=$(kubectl config view | grep server | cut -f 2- -d ":" | tr -d " ")
        addFullPathToCertsInKubeConfig
	cat $KUBECONFIG
        echo "Bluemix COnfig"
        cat ~/.bluemix/config.json
        sed -i "s/$OLD_CLUUSTER_NAME/$NEW_CLUSTER_NAME/" $MKPVYAML
        sed -i "s/$OLD_REQUEST_URL/$NEW_REQUEST_URL/" $MKPVYAML
        sed -i "s/$OLD_REGION/$NEW_REGION/" $YAMLPATH
	#make KUBECONFIGPATH=$KUBECONFIG PVG_PHASE=$PVG_PHASE armada-portworx-e2e-test | tee $E2E_PATH/log.txt
	make  PVG_PHASE=$PVG_PHASE armada-portworx-e2e-test | tee $E2E_PATH/log.txt
        exitStatus=$?
        ibmcloud ks cluster rm  --cluster $PVG_CLUSTER_CRUISER -f --force-delete-storage  
fi

echo "--- Cluster Details ---" >  $E2E_PATH/setupDetails.txt
echo "$CLUSTERDETAILS" >> $E2E_PATH/setupDetails.txt
echo "$CLUSTERDETAILS" >> $E2E_PATH/setupDetails.txt
echo "${PLUGINDETAILS}" >> $E2E_PATH/setupDetails.txt
echo "$CLUSTER_ID" >> $E2E_PATH/setupDetails.txt


echo "Finished ibmcloud block storage plugin e2e tests"
echo "In e2e script  existStatus = $exitStatus"
exit $exitStatus

