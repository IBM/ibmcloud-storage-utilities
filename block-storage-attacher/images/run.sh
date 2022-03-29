#!/bin/bash
# ******************************************************************************
# * Licensed Materials - Property of IBM
# * IBM Cloud Kubernetes Service, 5737-D43
# * (C) Copyright IBM Corp. 2022 All Rights Reserved.
# * US Government Users Restricted Rights - Use, duplication or
# * disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
# ******************************************************************************

set -ex

# The entry point for container
#VOLUME_CONFIG_FILE="/host/etc/iscsi-block-volume.conf"

# Copy block volume configurations
cp /home/armada-storage/iscsi-block-volume.conf /host/etc/

# Copy service files
mkdir -p /host/lib/ibmc-block-attacher
cp /home/armada-storage/iscsi-attach.sh /host/lib/ibmc-block-attacher/
cp /home/armada-storage/ibmc-block-attacher.service /host/lib/systemd/system/
ln -s -f /lib/systemd/system/ibmc-block-attacher.service /host/etc/systemd/system/multi-user.target.wants/ibmc-block-attacher.service

# Set the volume details from environment variables of pod to conf file
#sed -i 's/<iqn>/'$IQN'/' $VOLUME_CONFIG_FILE
#sed -i 's/<user>/'$USER'/' $VOLUME_CONFIG_FILE
#sed -i 's/<password>/'$PASSWORD'/' $VOLUME_CONFIG_FILE
#sed -i 's/<target>/'$TARGET'/' $VOLUME_CONFIG_FILE

/home/armada-storage/systemutil -action reload
/home/armada-storage/systemutil -target ibmc-block-attacher.service -action start
/home/armada-storage/block-storage-attacher

set +ex
