#!/bin/bash
# shellcheck disable=SC2162,SC2181,SC2086,SC2236,SC2162,SC2068,SC2002,SC2006,SC2001,SC2034,SC2069

ALLWORKERS="false"
NoOfWorkers=""

if [ $# -lt 1 ]; then
  echo "Worker IP is  not specified . The log will be collected form all available workers"
  ALLWORKERS="true"
elif [ $1 = --worker ]; then
   WORKER_IP=$2
   NoOfWorkers=$#
   echo "WORKER_IP = $WORKER_IP"
fi


NAMESPACE="ibm-system"
PXLOGS_PATH="/tmp/pxlogs/"
DS_NAME=""
JOB_NAME=""
DS_COMMAND="touch /tmp/pod/PXjournallog;journalctl -lu portworx* > /tmp/pod/PXjournallog;touch /tmp/pod/PXstatus;/opt/pwx/bin/pxctl  status > /tmp/pod/PXstatus;touch /tmp/pod/PXclusterstatus; /opt/pwx/bin/pxctl cluster provision-status > /tmp/pod/PXclusterstatus;/opt/pwx/bin/pxctl service diags -a;cp /var/cores/diags.tar.gz /tmp/pod"


JOB_COMMAND="touch /tmp/pod/PXjournallog;journalctl -lu portworx* > /tmp/pod/PXjournallog;sleep 1m;touch /tmp/pod/PXstatus;/opt/pwx/bin/pxctl  status > /tmp/pod/PXstatus;sleep 1m;touch /tmp/pod/PXclusterstatus; /opt/pwx/bin/pxctl cluster provision-status > /tmp/pod/PXclusterstatus;sleep 1m;/opt/pwx/bin/pxctl service diags -a;mv /var/cores/diags.tar.gz /tmp/pod;"

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
        image: "alpine:3.9"
        command:
          - sh
          - -c
          - nsenter -t 1 -m -u -i -n -p  -- bash -c "${DS_COMMAND}"
          #-  nsenter -t 1 -m -u -i -n -p; "touch /tmp/pod/PXstatus;/opt/pwx/bin/pxctl  status > /tmp/pod/PXstatus; touch /tmp/pod/PXclusterstatus; /opt/pwx/bin/pxctl cluster provision-status > /tmp/pod/PXclusterstatus;/opt/pwx/bin/pxctl service diags -a;cp /var/cores/diags.tar.gz /tmp/pod"
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
        image: "alpine:3.9"
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

sleep 15
kubectl -n $NAMESPACE get Pod -l app=runevery -l runevery=$DS_NAME -o jsonpath='{range .items[*]}{.metadata.name} {.spec.nodeName}{"\n"}{end}' | sort -k2 | while read podname nodename; do
        PXLOGS="${PXLOGS_PATH}/${nodename}"
        mkdir -p $PXLOGS
         kubectl  cp  ${NAMESPACE}/${podname}:/tmp/pod/PXstatus $PXLOGS/PXstatus -c tarry
         kubectl  cp  ${NAMESPACE}/${podname}:/tmp/pod/PXclusterstatus $PXLOGS/PXclusterstatus -c tarry
         kubectl  cp  ${NAMESPACE}/${podname}:/tmp/pod/PXjournallog $PXLOGS/PXjournallog -c  tarry
         kubectl  cp  ${NAMESPACE}/${podname}:/tmp/pod/diags.tar.gz $PXLOGS/diags.tar.gz -c tarry
done
}




deployJob () {
   echo " The workeres IP is specified. Collecting Logs from node $WORKER_IP"
   JOB_NAME=$(LC_CTYPE=C cat /dev/urandom | base64 | tr -dc a-z0-9 | fold -w 32 | head -n 1)
   (cat << EOF
apiVersion: batch/v1
kind: Job
metadata:
  name: ${JOB_NAME}
  namespace: ${NAMESPACE}
  labels:
    app: runon-shell
spec:
  template:
    spec:
      tolerations:
        - operator: "Exists"
      nodeSelector:
        kubernetes.io/hostname:  $WORKER_IP
      containers:
        - name: runon
          image: "alpine:3.10"
          command:
            - sh
            - -c
            - nsenter -t 1 -m -u -i -n -p  -- bash -c "${JOB_COMMAND}"
          securityContext:
            privileged: true
      hostPID: true
      restartPolicy: Never
EOF
) | if ! kubectl create -f - 2>&1 > /dev/null; then
  echo "unable to create job, bailing out"
  exit 1
fi

# get the uid
ID=$(kubectl get job ${JOB_NAME} -n ${NAMESPACE} -o 'jsonpath={.metadata.uid}')
if [ -z "${ID}" ]; then
  echo "ERR unable to get job id"
  exit 1
fi

sleep 20
POD=$(kubectl get pods -n ${NAMESPACE} -l controller-uid=${ID},job-name=${JOB_NAME} -o 'jsonpath={.items[].metadata.name}')
PXLOGS="${PXLOGS_PATH}/${WORKER_IP}"
        mkdir -p $PXLOGS
        echo "pod name = ${POD}"
         kubectl  cp  ${NAMESPACE}/${POD}:/tmp/pod/PXstatus $PXLOGS/PXstatus -c runon
         kubectl  cp  ${NAMESPACE}/${POD}:/tmp/pod/PXclusterstatus $PXLOGS/PXclusterstatus -c runon
         kubectl  cp  ${NAMESPACE}/${POD}:/tmp/pod/PXjournallog $PXLOGS/PXjournallog -c runon
         kubectl  cp  ${NAMESPACE}/${POD}:/tmp/pod/diags.tar.gz $PXLOGS/diags.tar.gz -c runon
}




if [[ $ALLWORKERS = true ]]; then
  deployDaemonSet
  sleep 20
  kubectl delete ds $DS_NAME -n $NAMESPACE
else
  for WORKER_IP in $@
  do
    if [ $WORKER_IP = --worker  ]; then
      continue
    fi
    deployJob
    sleep 20
    kubectl delete job $JOB_NAME -n $NAMESPACE
  done
fi
