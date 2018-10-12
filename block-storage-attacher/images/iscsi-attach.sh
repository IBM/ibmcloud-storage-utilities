#!/bin/bash
#/*******************************************************************************
# * Licensed Materials - Property of IBM
# * IBM Cloud Container Service, 5737-D43
# * (C) Copyright IBM Corp. 2017, 2018 All Rights Reserved.
# * US Government Users Restricted Rights - Use, duplication or
# * disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
# ******************************************************************************/

#set -ex
source /etc/iscsi-portworx-volume.conf

LOG=/var/log/ibmc-portworx-service.log
INITIATOR=/etc/iscsi/initiatorname.iscsi
ISCSI_CONF=/etc/iscsi/iscsid.conf
#ISCSIADM=/sbin/iscsiadm

if [ "$op" = "attach" ];
then
  if [ "$iqn" = "dummyiqn" ];
  then
    echo "`date`:Configuration parameters are not found." >> $LOG
    exit 0
  fi

  echo "`date`:====Start Attach====" >> $LOG

  if grep -q "^InitiatorName" $INITIATOR;
  then
    sed -i 's/^InitiatorName.*/InitiatorName='$iqn'/' $INITIATOR
  else
    echo "`date`:InitiatorName not found" >> $LOG
    echo "InitiatorName="$iqn >> $INITIATOR
  fi

  if grep -q "^node.session.auth.username" $ISCSI_CONF;
  then
    sed -i 's/^node.session.auth.username.*/node.session.auth.username = '$username'/' $ISCSI_CONF
  else
    echo "`date`:node.session.auth.username not found" >> $LOG
    echo "node.session.auth.username = "$username >> $ISCSI_CONF
  fi

  if grep -q "^node.session.auth.password" $ISCSI_CONF;
  then
    sed -i 's/^node.session.auth.password.*/node.session.auth.password = '$password'/' $ISCSI_CONF
  else
    echo "`date`:node.session.auth.password not found" >> $LOG
    echo "node.session.auth.password = "$password >> $ISCSI_CONF
  fi

  if grep -q "^discovery.sendtargets.auth.username" $ISCSI_CONF;
  then
    sed -i 's/^discovery.sendtargets.auth.username.*/discovery.sendtargets.auth.username = '$username'/' $ISCSI_CONF
  else
    echo "`date`:discovery.sendtargets.auth.username not found" >> $LOG
    echo "discovery.sendtargets.auth.username = "$username >> $ISCSI_CONF
  fi

  if grep -q "^discovery.sendtargets.auth.password" $ISCSI_CONF;
  then
    sed -i 's/^discovery.sendtargets.auth.password.*/discovery.sendtargets.auth.password = '$password'/' $ISCSI_CONF
  else
    echo "`date`:discovery.sendtargets.auth.password not found" >> $LOG
    echo "discovery.sendtargets.auth.password = "$password >> $ISCSI_CONF
  fi

  if grep -q "^node.startup" $ISCSI_CONF;
  then
    sed -i 's/^node.startup.*/node.startup = automatic/' $ISCSI_CONF
  else
    echo "`date`:node.startup not found" >> $LOG
    echo "node.startup = automatic" >> $ISCSI_CONF
  fi

  service iscsid restart
  service open-iscsi restart

  echo "`date`:iscsi discovery" >> $LOG
  iscsiadm -m discovery -t sendtargets -p $target_ip >> $LOG
  echo "`date`:iscsi login" >> $LOG
  iscsiadm -m node --login >> $LOG
  echo "`date`:iscsi rescan" >> $LOG
  iscsiadm -m session --rescan >> $LOG
  udevadm trigger
  rc=$?
  if [ $rc == 15 ]
  then
    rc=0
  fi

  echo "`date`:iscsi sessions" >> $LOG
  echo "`iscsiadm -m session`" >> $LOG

  found=false
  for devcounter in {1..60};
  do
    devs=`ls -l /dev/disk/by-path/ | grep "lun-$lunid " | wc -l`
    if [ $devs -ge 2 ];
    then
      echo "`date`:Found atleast 2 devices" >> $LOG
      found=true
      break
    fi
    iscsiadm -m session --rescan
    sleep 5
  done

  if [ $found == false ];
  then
    echo "`date`:Failed to attach the device" >> $LOG
    exit 1
  fi

#  echo "`date`:multipath -r" >> $LOG
#  echo "`multipath -r`" >> $LOG
  echo "`date`:multipath -ll" >> $LOG
  echo "`multipath -ll`" >> $LOG

  echo "`date`:Attached the volume successfully." >> $LOG

  found=false
  for pathcounter in {1..60};
  do
    paths=`multipathd show paths format "%w %i" | wc -l`
    mpaths=`multipathd show multipaths`
    if [ $paths -ge 2 -a "$mpaths" != "" ];
    then
      echo "`date`:Found the multipaths" >> $LOG
      found=true
      break
    fi
    sleep 5
  done

  if [ $found == false ];
  then
    echo "`date`:Failed to get the multipath" >> $LOG
    exit 1
  fi

  echo "`multipathd show paths format "%w %i"`" > /lib/ibmc-portworx/out_paths
  echo "`multipathd show multipaths`" > /lib/ibmc-portworx/out_multipaths
  exit $rc
elif [ "$op" = "detach" ];
then
  echo "`date`:====Start Detach====" >> $LOG
  echo "`echo 1 | tee /sys/block/$dm/slaves/*/device/delete`" >> $LOG
  multipath -f $mpath
  echo "`date`:Detached the volume successfully." >> $LOG
#  multipath -r
  exit 0
fi
