#!/bin/sh

set -ex

# The entry point for container
VOLUME_CONFIG_FILE="/host/etc/iscsi-portworx-volume.conf"

# Copy block volume configurations
cp /home/armada-storage/iscsi-portworx-volume.conf /host/etc/

# Copy service files
mkdir -p /host/lib/ibmc-portworx
cp /home/armada-storage/iscsi-attach.sh /host/lib/ibmc-portworx/
cp /home/armada-storage/ibmc-portworx.service /host/lib/systemd/system/
ln -s -f /lib/systemd/system/ibmc-portworx.service /host/etc/systemd/system/multi-user.target.wants/ibmc-portworx.service

# Set the volume details from environment variables of pod to conf file 
#sed -i 's/<iqn>/'$IQN'/' $VOLUME_CONFIG_FILE
#sed -i 's/<user>/'$USER'/' $VOLUME_CONFIG_FILE
#sed -i 's/<password>/'$PASSWORD'/' $VOLUME_CONFIG_FILE
#sed -i 's/<target>/'$TARGET'/' $VOLUME_CONFIG_FILE

/home/armada-storage/systemutil -action reload
/home/armada-storage/systemutil -target ibmc-portworx.service -action start
/home/armada-storage/block-storage-attacher

set +ex
