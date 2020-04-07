#!/bin/bash
# Licensed Materials - Property of IBM
#
# (C) Copyright IBM Corp. 2017 All Rights Reserved
#
# US Government Users Restricted Rights - Use, duplicate or
# disclosure restricted by GSA ADP Schedule Contract with
# IBM Corp.
# encoding: utf-8

SCRIPTS_FOLDER_PATH="src/github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher/tests/e2e/scripts"
E2EPATH="src/github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher/tests/e2e/e2e-tests/"
SCRIPTS_FOLDER_PATH="$GOPATH/$SCRIPTS_FOLDER_PATH"
E2E_PATH="$GOPATH/$E2EPATH"


export ISSUE_REPO="${DEFAULT_ISSUE_REPO}"

function set_issue_repo {
    # Set issue_repo to $1
    set +x
    ISSUE_REPO="${1}"
    set -x
}

function write_issue_repo {
    # Called by trap on EXIT; write the searchable to the console (log)
    set +x
    echo "GHE_ISSUE_REPO=${ISSUE_REPO}"
}

trap write_issue_repo EXIT

function setKubeConfig {
    if [ -z $1 ]; then
        echo "Cluster name not specified, ${FUNCNAME[0]} skipped"
        return 0
    fi
    set_issue_repo "armada-api"

    # Get the kube config from the `bx` cli and export KUBECONFIG for
    # your current bash session

    cluster_name=$1
    echo "Generating Kube Config through 'bx cs cluster-config $cluster_name --admin' and exporting KUBECONFIG"
    configfile=$(ibmcloud ks  cluster-config $cluster_name --admin | grep ^export | cut -d '=' -f 2)
    cat $configfile
    export KUBECONFIG=$configfile

    test $KUBECONFIG
    set_issue_repo ${DEFAULT_ISSUE_REPO}

}

function slack_commentary {
    # NOTE(cjschaef): we don't want to perform any slack comment calls
    echo "Ignoring 'slack comment' call"
    return

    # If not specified, indent 5 spaces
    if [ -z ${GATE_TEST_NAME} ]; then
        GATE_TEST_NAME="     "
    fi
    /dove/pvg_slack_message.py --phase $PVG_PHASE --commentary "$GATE_TEST_NAME - $1" || true
}

function print_info {
    echo "============== $(date -u +'%D %T %Z') =================="
    echo "  $1"
    echo "======================================================="
}

function cruiser_create {
    # Parameters:
    # cluster_name
    # machine_type
    # number_of_workers
    cluster_create cruiser $@
}

function patrol_create {
    # Parameters:
    # cluster_name
    if [ -z $1 ]; then
        echo "Cluster name not specified, ${FUNCNAME[0]} skipped"
        return 0
    fi
    cluster_create patrol $@
}

# Create a cluster
function cluster_create {
    set_issue_repo "armada-api"
    cluster_type=$1
    cluster_name=$2
    if [[ $cluster_type == "patrol" ]]; then
        ibmcloud ks cluster-create --name $cluster_name
    elif [[ $cluster_type == "cruiser" ]]; then
        machine_type=$3
        cluster_workers=$4
        eval PVG_CRUISER_PUBLIC_VLAN=\$${PVG_CLUSTER_LOCATION}_PVG_CRUISER_PUBLIC_VLAN
        eval PVG_CRUISER_PRIVATE_VLAN=\$${PVG_CLUSTER_LOCATION}_PVG_CRUISER_PRIVATE_VLAN
        if [[ -n $PVG_CLUSTER_KUBE_VERSION ]]; then
	        ibmcloud ks  cluster-create --name $cluster_name --location $PVG_CLUSTER_LOCATION \
	            --public-vlan $PVG_CRUISER_PUBLIC_VLAN --private-vlan $PVG_CRUISER_PRIVATE_VLAN \
	            --workers $cluster_workers --machine-type $machine_type --kube-version $PVG_CLUSTER_KUBE_VERSION
        else
	        ibmcloud ks cluster-create --name $cluster_name --location $PVG_CLUSTER_LOCATION \
	            --public-vlan $PVG_CRUISER_PUBLIC_VLAN --private-vlan $PVG_CRUISER_PRIVATE_VLAN \
	            --workers $cluster_workers --machine-type $machine_type
        fi
    else
        echo "type must be set to one of 'patrol' or 'cruiser'"
        exit 1
    fi
    set_issue_repo ${DEFAULT_ISSUE_REPO}
}

# Wait for cluster delete
function check_cluster_deleted {
    if [ -z $1 ]; then
        echo "Cluster name not specified, ${FUNCNAME[0]} skipped"
        return 0
    fi
    attempts=0
    set_issue_repo "armada-deploy"
    set +x
    cluster_name=$1
    cluster_id=$(ibmcloud ks clusters | awk "/$cluster_name/"'{print $2}')
    echo "Waiting for $cluster_name ($cluster_id) to be deleted..."
    while true; do
        cluster_count=$(ibmcloud ks clusters | grep $cluster_name | wc -l | xargs)
        echo "$cluster_count instances still exist"
        if [[ $cluster_count -eq 0 ]]; then
            break;
        fi
        state=$(ibmcloud ks clusters | awk "/$cluster_name/"'{print $3}')

        attempts=$((attempts+1))
        if [[ $state == "*_failed" ]]; then
            echo "$cluster_name ($cluster_id) entered a $state state. Exiting"
            slack_commentary "$cluster_name ($cluster_id) entered a $state state."
            exit 1
        fi
        if [[ $attempts -gt 120 ]]; then
            echo "$cluster_name ($cluster_id) failed to be deleted after 15 minutes. Exiting."
            slack_commentary "$cluster_name ($cluster_id) failed to be deleted after 15 minutes."
            # Show cluster workers state as it is helpful.
            ibmcloud ks workers $cluster_name
            exit 2
        fi
        echo "$cluster_name ($cluster_id) state == $state.  Sleeping 30 seconds"
        slack_commentary "$cluster_name ($cluster_id) state = $state. Sleeping 30 seconds."
        sleep 30
    done
    ibmcloud ks clusters
    set -x
    set_issue_repo ${DEFAULT_ISSUE_REPO}
}

# Check cluster state
function check_cluster_state {
    if [ -z $1 ]; then
        echo "Cluster name not specified, ${FUNCNAME[0]} skipped"
        return 0
    fi
    attempts=0
    set_issue_repo "armada-deploy"
    set +x
    cluster_name=$1
    cluster_id=$(ibmcloud ks clusters | awk "/$cluster_name/"'{print $2}')
    echo "Waiting for $cluster_name ($cluster_id) to reach deployed state..."
    while true; do
        state=$(ibmcloud ks clusters | awk "/$cluster_name/"'{print $3}')
        attempts=$((attempts+1))
        if [[ $state == "*_failed" ]]; then
            echo "$cluster_name ($cluster_id) entered a $state state. Exiting"
            slack_commentary "$cluster_name ($cluster_id) entered a $state state."
            exit 1
        # There are multiple states that equate to deployed if $state matches
        # any of them, then break out of the loop.  Without the $, normals would
        # be a valid match.
        elif [[ ${state} =~ deployed$|normal$|warning$|critical$|pending$ ]]; then
            echo "$cluster_name ($cluster_id) reached a valid state!"
            slack_commentary "$cluster_name ($cluster_id) reached a valid state!"
            break
        fi
        if [[ $attempts -gt 120 ]]; then
            echo "$cluster_name ($cluster_id) failed to reach a valid state after 15 minutes. Exiting."
            slack_commentary "$cluster_name ($cluster_id) failed to reach a valid state after 15 minutes."
            # Show cluster workers state as it is helpful.
            ibmcloud ks workers $cluster_name
            exit 2
        fi
        echo "$cluster_name ($cluster_id) state == $state.  Sleeping 30 seconds"
        slack_commentary "$cluster_name ($cluster_id) state = $state. Sleeping 30 seconds."
        sleep 30
    done
    ibmcloud ks clusters
    set -x
    set_issue_repo ${DEFAULT_ISSUE_REPO}
}

# Check cluster worker state
function check_worker_state {
    if [ -z $1 ]; then
        echo "Cluster name not specified, ${FUNCNAME[0]} skipped"
        return 0
    fi
    all_workers_good=0
    set_issue_repo "armada-deploy"
    set +x
    cluster_name=$1

    TIMEOUT=${WORKER_STATE_TIMEOUT:-90}
    echo "Waiting for $cluster_name workers to reach deployed state( $TIMEOUT minutes )..."
    slack_commentary "Waiting for $cluster_name workers to reach deployed state..."
    set +e
    ibmcloud ks workers $cluster_name | grep $PVG_CLUSTER_LOCATION
    set -e

    # Try for up to 90 minutes(default) for the workers to reach deployed state
    for ((i=1; i<=TIMEOUT; i++)); do
        oldifs="$IFS"
        IFS=$'\n'
        workers=($(ibmcloud ks workers $cluster_name | grep $PVG_CLUSTER_LOCATION))
        IFS="$oldifs"
        worker_cnt=${#workers[@]}
        # Inspect the state of each worker
        for worker in "${workers[@]}"; do
            # Fail if any are in failed state
            state=$(echo $worker | awk '{print $5}')
            worker_id=$(echo $worker | awk '{print $1}')
            if [[ $state == "*_failed" ]]; then
                if [[ $state == "bootstrap_failed" ]]; then
                    set_issue_repo "armada-bootstrap"
                elif [[ $state == "provision_failed" ]]; then
                    set_issue_repo "armada-cluster"
                fi
                echo "$cluster_name worker $worker_id entered a $state state. Exiting"
                slack_commentary "$cluster_name worker $worker_id entered a $state state."
                exit 1
            elif [[ ${state} =~ deployed$|normal$|warning$|critical$ ]]; then
                echo "$cluster_name worker $worker_id state == $state."
                slack_commentary "$cluster_name worker $worker_id state == $state."
                # Count the number of workers in deployed state
                all_workers_good=$((all_workers_good+1))
            fi
        done
        if [[ $worker_cnt -eq $all_workers_good ]]; then
            # Break out of the 30 minute loop since all workers reach deployed state
            break
        else
            # Else sleep 60 seconds

            # Ignore failures on this command call
            ibmcloud ks workers $cluster_name || true
            status_msg="$all_workers_good of $worker_cnt $cluster_name workers are in deployed state. Sleeping 60 seconds."
            echo "$status_msg"
            slack_commentary "$status_msg"
            sleep 60
            all_workers_good=0
        fi
    done

    # Ignore failures on this command call
    ibmcloud ks workers $cluster_name || true
    if [[ $worker_cnt -ne $all_workers_good ]]; then
        # 30 minutes have passed and not all workers reached deployed state
        # so return as failed.
        workers=($(ibmcloud ks workers $cluster_name | grep $PVG_CLUSTER_LOCATION))
        for worker in "${workers[@]}"; do
            state=$(echo $worker | awk '{print $5}')
            if [[ $state == "bootstrapping" ]]; then
                set_issue_repo "armada-bootstrap"
            elif [[ $state == "provisioning" || $state == "pending_provision" ]]; then
                set_issue_repo "armada-cluster"
            fi
        done
        echo "Not all $cluster_name workers reached deployed state in 40 minutes."
        slack_commentary "Not all $cluster_name workers reached deployed state in 40 minutes."
        return 1
    fi
    set -x
    set_issue_repo ${DEFAULT_ISSUE_REPO}
    # All is good
    return 0
}

function rm_cluster {
    if [ -z $1 ]; then
        echo "Cluster name not specified, ${FUNCNAME[0]} skipped"
        return 0
    fi
    removed=1
    cluster_name=$1

    ibmcloud ks clusters

    for i in {1..3}; do
        if ibmcloud ks cluster-rm $cluster_name -f; then
            sleep 30
            # Remove old kubeconfig files aswell
            rm -rf $HOME/.bluemix/plugins/container-service/clusters/$cluster_name
            removed=0
            break
        fi
        sleep 30
    done
    return $removed
}

function source_armada_bom {
    if [ -f /tmp/ansible_bom.rc ]; then
        # Source the rc file to have the images as environment variables
        . /tmp/ansible_bom.rc
    else
        echo "FAIL! The `/tmp/ansible_bom.rc` does not exist"
    fi
}

function setKubeConfigLocal {
    # Setup the kubeconfig based off the below location
    kube_config_location="/armada-gate/etc/admin-kubeconfig"

    echo "$kube_config_location"
    export KUBECONFIG="${kube_config_location}"
    # NOTE(tjcocozz): These environment variables are inherited into our gate container
    # from the ansible-deploy/base kubernetes on the carrier and are setting the default
    # namespace to `armada` which doesn't exist on the customer kubernetes deployment.
    # We must unset these vars to allow kubernetes set the namespace to `default`
    unset KUBERNETES_PORT
    unset KUBERNETES_PORT_443_TCP_PORT
    unset KUBERNETES_SERVICE_HOST
}

function docker_reg_login {
    echo 'Logging Into Docker Registry'
    docker version
    docker login -u token -p $INTERNAL_REGISTRY_TOKEN $REGISTRY_LOCATION
    docker images
}

function bx_login {
    echo 'Logging Into BlueMix Containers service'
    ibmcloud --version
    ibmcloud plugin list

    ibmcloud login -a $PVG_BX_DASH_A -u $PVG_BX_USER -p $PVG_BX_PWD -c $PVG_BX_DASH_C -r $ARMADA_REGION
    #bx cs init --host $ARMADA_API_ENDPOINT

}

function setupKubernetes {
    # NOTE(cjschaef): we want to prevent using the IBM built kubectl
    echo "Ignoring install of IBM built kubectl"
    return

    echo "Use Kubectl binary from kubernetes build"

    kube_tests_image=$REGISTRY_LOCATION/armada-master/kube_tests:$KUBE_TEST_IMAGE_VERSION
    docker pull $kube_tests_image

    # Copy the source and binary tar balls out of the test container into /kubernetes
    container_id=$(docker run -d --entrypoint sleep $kube_tests_image 300)
    docker export $container_id | tar -x k8s
    pushd k8s
    ls -la
    tar -zxf kubernetes-source-*.tar.gz -C /
    tar -zxf kubernetes-build-*.tar.gz -C /kubernetes
    chmod +x /kubernetes/_output/dockerized/bin/linux/amd64/*
    popd

    cp /kubernetes/_output/dockerized/bin/linux/amd64/kubectl /usr/local/bin/
    kubectl version --client
}

function addFullPathToCertsInKubeConfig {
    # The e2e tests expect the full path of the certs
    # to be in the kube confg. Prior to calling this function.
    # it is expected to have a `KUBECONFIG` variable defined

    CONFIG_DIR=$(dirname $KUBECONFIG)
    pushd ${CONFIG_DIR}

    for certFile in $(ls | grep -E ".*.pem"); do
        certFilePATH=${CONFIG_DIR}/${certFile}
        # Replace the certs with full path unless they already have the full path
        sed "s|[^\/]$certFile| $certFilePATH|g" $KUBECONFIG > /tmp/kubeconfig.yml;mv /tmp/kubeconfig.yml $KUBECONFIG
    done
    popd
}

# Check if the POD has reached running state (timeout: 300sec)
function check_pod_state {
  attempts=0
  pod_name=$1
  while true; do
    attempts=$((attempts+1))
      pod_status=$(kubectl get pods -n kube-system | awk "/$pod_name/"'{print $2}')
      readarray -t lines < <(echo "$pod_status")
         if [ "${lines[0]}" = "1/1" ]; then
            echo "$pod_name is  running ."
            break
         fi
    if [[ $attempts -gt 30 ]]; then
      echo "$pod_name  were not running well."
      kubectl get pods -n kube-system| awk "/$pod_name-/"
      exit 1
    fi
    echo "$pod_name state == $pod_status  Sleeping 10 seconds"
    sleep 10
  done
}
function check_portworx_pod_state {
  attempts=0
  pod_name=$1
  while true; do
    attempts=$((attempts+1))
      pod_status=$(kubectl get pods -n kube-system | awk "/$pod_name/"'{print $2}')
      readarray -t lines < <(echo "$pod_status")
         if [ "${lines[0]}" = "1/1" ]; then
            echo "$pod_name is  running ."
            break
         fi
    if [[ $attempts -gt 50 ]]; then
      echo "$pod_name  were not running well."
      kubectl get pods -n kube-system| awk "/$pod_name-/"
      exit 1
    fi
    echo "$pod_name state == $pod_status  Sleeping 10 seconds"
    sleep 30
  done
}


# Check if the Deployment has reached running state (timeout: 300sec)
function check_deployment_state {
  attempts=0
  deployment_name=$1
  while true; do
    attempts=$((attempts+1))
    deployment_status=$(kubectl get pods -n kube-system | awk "/$deployment_name/"'{print $1}')
    if [   "$deployment_status" = "1" ]; then
      echo "$deployment_name is  running ."
      break
    fi
    if [[ $attempts -gt 30 ]]; then
      echo "$deployment_name  were not running well."
      kubectl get pods -n kube-system| awk "/$deployment_name/"
      exit 1
    fi
    echo "$deployment_name state == $deployment_status  Sleeping 10 seconds"
    sleep 10
  done
}

# Check if the check_daemonset_state has reached running state (timeout: 300sec)
# and desired matches with available
function check_daemonset_state {
  attempts=0
  ds_name=$1
  while true; do
    attempts=$((attempts+1))
    ds_status_desired=$(kubectl get ds -n kube-system | awk "/$ds_name/"'{print $2}')
    ds_status_available=$(kubectl get ds -n kube-system | awk "/$ds_name/"'{print $5}')
    if [   "$ds_status_desired" = "$ds_status_available" ]; then
      echo "$ds_name is  running and available ds instances: $ds_status_available"
      break
    fi
    if [[ $attempts -gt 30 ]]; then
      echo "$ds_name  were not running well. Instances Desired:$deployment_status_desired, Instances Available:$deployment_status_available"
      kubectl get deployment -n kube-system| awk "/$ds_name/"
      exit 1
    fi
    echo "DS:$ds_name, Desired:$ds_status_desired, Available:$ds_status_available  Sleeping 10 seconds"
    sleep 10
  done
}

# Install/Upgrade blockvlome-attacher helm chart
function install_blockvolume_plugin {
	if [ -z $HELM_CHART ]; then
        echo "helm chart not found. Hence exiting"
        exit 1
    fi
    echo "Installing helm chart ibm-block-storage-attacher .."
	# INSTALL HELM TILLER (Attempt again, if already installed)
	echo "Initialize tiller AND Wait till running"
	helm init --force-upgrade
	check_pod_state "tiller-deploy"

	# INSTALL/UPGRADE HELM CHART
	helm_values_override="--set image.repository=$IMAGE_REGISTRY/$USER_NAMESPACE/$PLUGIN_IMAGE --set image.build=$PLUGIN_BUILD"
	helm_install_cmd="helm install $helm_values_override $HELM_CHART"

	# CHECK FOR UPGRADE

        echo "Checking for existing helm chart ibm-block-storage-attacher on cluster .."
        helm_release=$(helm ls | grep DEPLOYED | awk "/ibm-block-storage-attacher/"'{print $1}')
        if [   "$helm_release" != "" ]; then
          echo "Existing release $helm_release found for chart ibm-block-storage-attacher"
          helm_install_cmd="helm upgrade --force --recreate-pods $helm_values_override $helm_release $HELM_CHART"
        fi

	# DO HELM INSTALLATION
	echo "Executing: $helm_install_cmd"
	set +e
	for i in {1..5}; do
	    if $helm_install_cmd ; then
	        echo "helm install started"
	        break
	    fi
	    sleep 30
	done
	set -e
	echo "Checking storage plugin and driver status and wait till running"
}
function install_portworx_plugin {
        touch $E2E_PATH/debug.txt
        echo "Debug" > $E2E_PATH/debug.txt
	if [ -z $PORTWORX_HELM_CHART ]; then
        echo "port worx helm chart not found. Hence exiting" >> $E2E_PATH/debug.txt
        exit 1
    fi
    echo "Installing helm chart portworx plugin  .."
	# INSTALL HELM TILLER (Attempt again, if already installed)
	echo "Initialize tiller AND Wait till running" >> $E2E_PATH/debug.txt
	helm init --force-upgrade
	check_pod_state "tiller-deploy"
	
        # INSTALL/UPGRADE HELM CHART
	helm_values_override="--set kvdb=$ETCD_SET1,clusterName=$(uuidgen),usedrivesAndPartitions=true,usefileSystemDrive=true,imageVersion=$PORTWORXVER,secretType=$IBMSECRETTYPE,drives=none,etcd.secret=$PXETCDSECRET"
        echo "Helm installtion start" >> $E2E_PATH/debug.txt
	helm_install_cmd="helm install $helm_values_override $PORTWORX_HELM_CHART --name portworx"
        echo "Helm installtion done" >> $E2E_PATH/debug.txt
        
	# CHECK FOR UPGRADE

        echo "Checking for existing helm chart ibm-block-storage-attacher on cluster .."
        portworx_helm_release=$(helm ls | grep DEPLOYED | awk "/portworx/"'{print $1}')
        if [   "$portworx_helm_release" != "" ]; then
          echo "Existing release $portworx_helm_release found for chart ibm-block-storage-attacher"
          return  
        fi
	# DO HELM INSTALLATION
	echo "Executing: $helm_install_cmd"
	set +e
	for i in {1..5}; do
	    if $helm_install_cmd ; then
	        echo "helm install started" >> $E2E_PATH/debug.txt
	        break
	    fi
	    sleep 30
	done
	set -e
	echo "Checking portworx status and wait till running" >> $E2E_PATH/debug.txt
}

