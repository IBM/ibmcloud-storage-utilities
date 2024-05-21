#!/bin/bash
# ******************************************************************************
# * Licensed Materials - Property of IBM
# * IBM Cloud Kubernetes Service, 5737-D43
# * (C) Copyright IBM Corp. 2022 All Rights Reserved.
# * US Government Users Restricted Rights - Use, duplication or
# * disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
# ******************************************************************************

# shellcheck disable=SC2162,SC2181,SC2086,SC2236,SC2162,SC2068,SC2002,SC2006,SC2001,SC2034,SC2069

PX_NAMESPACE=$1
ALLWORKERS="true"

NAMESPACE="kube-system"
PXLOGS_PATH="/tmp/pxlogs/"
DS_NAME=""

DS_NAME=$(LC_CTYPE=C cat /dev/urandom | base64 | tr -dc a-z0-9 | fold -w 32 | head -n 1)

deployDaemonSet () {

(cat << EOF
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: ${DS_NAME}
  namespace: ${NAMESPACE}
  labels:
    app: runevery
  annotations:
    client: "$(hostname)"
    srctime: "${SRCTIME}"
    #logfile: "${LOGFILE}"
spec:
  selector:
    matchLabels:
      app: runevery
      runevery: ${DS_NAME}
  template:
    metadata:
      labels:
        app: runevery
        runevery: ${DS_NAME}
      annotations:
        client: "$(hostname)"
        srctime: "${SRCTIME}"
        #logfile: "${LOGFILE}"
    spec:
      tolerations:
      - operator: "Exists"
      hostNetwork: true
      initContainers:
      - name: doit
        image: "alpine:3.18"
        command:
          - sh
          - -c
          - nsenter -t 1 -m -u -i -n -p  -- bash -c "touch /tmp/pod/PXjournallog;journalctl -lu portworx* > /tmp/pod/PXjournallog;touch /tmp/pod/pxkvdbstatus;/opt/pwx/bin/pxctl service kvdb members > /tmp/pod/pxkvdbstatus;touch /tmp/pod/PXstatus;/opt/pwx/bin/pxctl  status > /tmp/pod/PXstatus;touch /tmp/pod/pxsvdumps;pxctl sv dump --nodestats > /tmp/pod/pxsvdumps;touch /tmp/pod/PXclusterstatus; /opt/pwx/bin/pxctl cluster provision-status > /tmp/pod/PXclusterstatus;rm -rf /tmp/pod/*.tar.gz;/opt/pwx/bin/pxctl service diags -a -o /tmp/pod/diags.tar.gz > /tmp/pod/diags_logs;cat tmp/pod/diags_logs"
          #-  nsenter -t 1 -m -u -i -n -p; "touch /tmp/pod/PXstatus;/opt/pwx/bin/pxctl  status > /tmp/pod/PXstatus; touch /tmp/pod/PXclusterstatus; /opt/pwx/bin/pxctl cluster provision-status > /tmp/pod/PXclusterstatus;/opt/pwx/bin/pxctl service diags -a;cp /var/cores/diags.tar.gz /tmp/pod/"
        securityContext:
          privileged: true
        volumeMounts:
        - name: tmp-pod
          mountPath: /tmp/pod
      hostPID: true
      volumes:
      - name: host-core
        hostPath:
           path: /var/cores
      containers:
      - name: tarry
        image: "alpine:3.18"
        command:
          - sh
          - -c
          - while true; do sleep 15; done
        securityContext:
          privileged: true
        volumeMounts:
        - name: tmp-pod
          mountPath: /tmp/pod
      hostPID: true
      volumes:
      - name: host-core
        hostPath:
           path: /var/cores
      - name: tmp-pod
        hostPath:
           path: /tmp/pod
EOF
) | if ! kubectl create -f - 2>&1 > /dev/null; then
  echo "unable to create DaemonSet, bailing out"
  exit 1
fi

sleep 300
kubectl -n $NAMESPACE get Pod -l app=runevery -l runevery=$DS_NAME -o jsonpath='{range .items[*]}{.metadata.name} {.spec.nodeName}{"\n"}{end}' | sort -k2 | while read podname nodename; do
        PXLOGS="${PXLOGS_PATH}/${nodename}"
        mkdir -p $PXLOGS
        diagsfile=$(kubectl logs ${podname} -n ${NAMESPACE} -c doit | grep -i "Generated diags" | awk '{print $3}' | awk -F '/' '{print $4}')
        kubectl -n ${PX_NAMESPACE} logs -l name=stork --tail=99999 > $PXLOGS/stork.log
        kubectl  cp  ${NAMESPACE}/${podname}:/tmp/pod/pxsvdumps $PXLOGS/pxsvdumps -c tarry --retries 10
         kubectl  cp  ${NAMESPACE}/${podname}:/tmp/pod/PXstatus $PXLOGS/PXstatus -c tarry --retries 10
         kubectl  cp  ${NAMESPACE}/${podname}:/tmp/pod/PXclusterstatus $PXLOGS/PXclusterstatus -c tarry --retries 10
         kubectl  cp  ${NAMESPACE}/${podname}:/tmp/pod/PXjournallog $PXLOGS/PXjournallog -c  tarry --retries 10
         kubectl  cp  ${NAMESPACE}/${podname}:/tmp/pod/${diagsfile} $PXLOGS/diags.tar.gz -c tarry --retries 10
         kubectl  cp  ${NAMESPACE}/${podname}:/tmp/pod/pxkvdbstatus $PXLOGS/pxkvdbstatus -c  tarry --retries 10
done
}

  deployDaemonSet
  sleep 20
  kubectl delete ds $DS_NAME -n $NAMESPACE
