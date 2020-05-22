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
export PVG_BX_MACHINE_TYPE="u2c.2x4"
export PVG_BX_CLUSTER_WORKERS_COUNT=1
export PVG_BX_DASH_O="contsto2@in.ibm.com"

# Setup the region data
if [[ "${TEST_REGION}" == "armada-dev" || "${TEST_REGION}" == "armada-prestage" || "${TEST_REGION}" == "armada-stage" ]]; then
    bx cs region-set "us-south"
else
    bx cs region-set "${TEST_REGION}"
fi

# Return the first zone for cluster creation
function get_first_zone {
    local zones_array=( `bx cs zones |tail -n +3` )      
    echo "${zones_array[0]}"
}

# Return public VLAN of the zone
function get_public_vlan {
    local public_vlan=( `bx cs vlans --zone $1 |grep -i Cruiser|grep -i public | head -n 1|awk '{ print $1}'` )      
    echo "$public_vlan"
}

# Return private VLAN of the zone
function get_private_vlan {
    local private_vlan=( `bx cs vlans --zone $1 |grep -i Cruiser|grep -i private | head -n 1|awk '{ print $1}'` )      
    echo "$private_vlan"
}

# Adding worker pools and zones_array
function add_worker_pools_and_zones {
    set -x
    local zones_array=( `bx cs zones |tail -n +3` )
    for zone in "${zones_array[@]}"
    do
        bx cs worker-pool-create --name pool_$zone --cluster $1 --flavor $PVG_BX_MACHINE_TYPE --size-per-zone  $PVG_BX_CLUSTER_WORKERS_COUNT
        public_zone=$(get_public_vlan $zone)
        private_zone=$(get_private_vlan $zone)
        bx cs zone-add --zone $zone --cluster $1 --worker-pools pool_$zone --public-vlan $public_zone --private-vlan $private_zone
    done
    set +x
}

# Remove the created clusters
function rm_cluster {
    removed=1
    cluster_name=$1

    bx cs clusters
    set +e
    for i in {1..3}; do
        if bx cs cluster-rm $cluster_name -f; then
            removed=0
            break
        fi
        sleep 30
    done

    return $removed
}

# Check for cluste update completion
function check_cluster_update {
    if [ -z $1 ]; then
        echo "Cluster name not specified, ${FUNCNAME[0]} skipped"
        return 0
    fi

    not_updated=1
    cluster_name=$1
    expected_kube_version=$2
    bx cs clusters
    set +e
    set -x
    for i in {1..30}; do
        local kube_version=( `bx cs clusters | grep $cluster_name |awk '{ print $9}'` )
        if [[ "$kube_version" =~ $expected_kube_version ]]; then
            sleep 60
            not_updated=0
            break
        fi
        sleep 60
    done

    return $not_updated
}

# Setup the region data
if [[ "${TEST_REGION}" == "armada-dev" || "${TEST_REGION}" == "armada-prestage" || "${TEST_REGION}" == "armada-stage" ]]; then
    export PVG_BX_DASH_C=8ee729d7f903db130b00257d91b6977f
    export PVG_BX_DASH_A=https://api.stage1.ng.bluemix.net
    if [[ "${TEST_REGION}" == "armada-dev" ]]; then
        export ARMADA_API_ENDPOINT=https://dev.cont.bluemix.net
    elif [[ "${TEST_REGION}" == "armada-prestage" ]]; then
        export ARMADA_API_ENDPOINT=https://prestage.cont.bluemix.net
    else
        export ARMADA_API_ENDPOINT=https://stage.cont.bluemix.net
    fi
else
    export PVG_BX_DASH_C=e242f140687cd68a8e037b26680e0f04
    if [[ "${TEST_REGION}" == "us-south" ]]; then
        export PVG_BX_DASH_A=https://api.ng.bluemix.net
        export ARMADA_API_ENDPOINT=https://us-south.containers.bluemix.net
    elif [[ "${TEST_REGION}" == "us-east" ]]; then
        export PVG_BX_DASH_A=https://api.us-east.bluemix.net
        export ARMADA_API_ENDPOINT=https://us-east.containers.bluemix.net
    elif [[ "${TEST_REGION}" == "uk-south" ]]; then
        export PVG_BX_DASH_A=https://api.eu-gb.bluemix.net
        export ARMADA_API_ENDPOINT=https://uk-south.containers.bluemix.net
    elif [[ "${TEST_REGION}" == "eu-central" ]]; then
        export PVG_BX_DASH_A=https://api.eu-de.bluemix.net
        export ARMADA_API_ENDPOINT=https://eu-central.containers.bluemix.net
    elif [[ "${TEST_REGION}" == "ap-south" ]]; then
        export PVG_BX_DASH_A=https://api.ng.bluemix.net
        export ARMADA_API_ENDPOINT=https://ap-south.containers.bluemix.net
    elif [[ "${TEST_REGION}" == "ap-north" ]]; then
        export PVG_BX_DASH_A=https://api.ng.bluemix.net
        export ARMADA_API_ENDPOINT=https://ap-north.containers.bluemix.net
    fi
fi
