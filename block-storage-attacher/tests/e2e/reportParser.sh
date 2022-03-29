#!/bin/bash
# ******************************************************************************
# * Licensed Materials - Property of IBM
# * IBM Cloud Kubernetes Service, 5737-D43
# * (C) Copyright IBM Corp. 2022 All Rights Reserved.
# * US Government Users Restricted Rights - Use, duplication or
# * disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
# ******************************************************************************

# shellcheck disable=SC1009,SC1079,SC1073,SC1072,SC1078

### Build the sample Test Report
set -x
echo "BlockVolumeAttacherTest: Cluster-Creation: NA" >finalReport.txt
echo "BlockVolumeAttacherTest: Plugin-Installation: NA" >>finalReport.txt
echo "BlockVolumeAttacherTest: PVC-Creation: NA" >>finalReport.txt
echo "BlockVolumeAttacherTest: Write-Volume: NA" >>finalReport.txt
echo "BlockVolumeAttacherTest: Read-Volume: NA" >>finalReport.txt
echo "BlockVolumeAttacherTest: Delete-PVC: NA" >>finalReport.txt
echo "BlockVolumeAttacherTest: MZ-PVC-Creation: NA" >>finalReport.txt
echo "BlockVolumeAttacherTest: MZ-POD-Creation: NA" >>finalReport.txt
echo "BlockVolumeAttacherTest: MZ-Write-Volume: NA" >>finalReport.txt
echo "BlockVolumeAttacherTest: MZ-Read-Volume: NA" >>finalReport.txt
echo "BlockVolumeAttacherTest: MZ-Force-Volume-Reschedule: NA" >>finalReport.txt
echo "BlockVolumeAttacherTest: MZ-Read-Volume-After-Reschedule: NA" >>finalReport.txt

####

E2E_PATH="$GOPATH/src/github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher/tests/e2e/e2e-tests/

# E2E binary execution logs are saved
# in log.txt and cluster setup plus multizone logs
# are saved in testList.txt

logFile=$E2E_PATH/log.txt
e2eFile=$E2E_PATH/e2eTests.txt

setupTests="Cluster-Creation,Plugin-Installation"
e2eTests=$(cat $logFile | grep -oP '(?s)(?<=\[TEST-LIST-BEGIN\]).*(?=\[TEST-LIST-END\])'| head -1)

if  [  -e $e2eFile ]
then

    for i in $(echo $setupTests | sed "s/,/ /g")
    do
        echo $i
        grep -w "BlockVolumeAttacherTest: $i"  $e2eFile
        retVal=$?
        if [ $retVal -ne 0 ]; then
            sed -i "s/BlockVolumeAttacherTest: $i: .*$/BlockVolumeAttacherTest: $i: FAIL/g" finalReport.txt
            break
        else
            sed -i "s/BlockVolumeAttacherTest: $i: .*$/BlockVolumeAttacherTest: $i: PASS/g" finalReport.txt
        fi
    done
fi

if  [  -e $logFile ]
then
    for i in $(echo $e2eTests | sed "s/,/ /g")
    do
        grep -w "BlockVolumeAttacherTest: $i"  $logFile
        retVal=$?

        echo "return value" $retVal
        if [ $retVal -ne 0 ]; then
            sed -i "s/BlockVolumeAttacherTest: $i: NA/BlockVolumeAttacherTest: $i: FAIL/g" finalReport.txt
            break
        else
            sed -i "s/BlockVolumeAttacherTest: $i: NA/BlockVolumeAttacherTest: $i: PASS/g" finalReport.txt
        fi
    done
fi

set +x
