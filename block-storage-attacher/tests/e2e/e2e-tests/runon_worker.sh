#!/bin/bash
# shellcheck disable=SC2068,SC2034,SC2124,SC2069
#. ./connect_worker.sh  worker private ip bash
if [ $# -lt 3 ]; then
  echo "runon [node IP] [Job_Name] [command]"
  exit 1;
fi

function cleanup {
  rm -rf "${TEMP_DIR}"
}

function pod_status {
  kubectl get pods "$1" -n "${NAMESPACE}" -o json | jq '.status.conditions[] | select(.type == "Ready") | .status ' | sed 's/\"//g'
}

# create a random pod name
#JOB_NAME=$(LC_CTYPE=C cat /dev/urandom | base64 | tr -dc a-z0-9 | fold -w 32 | head -n 1)
JOB_NAME=$2
NAMESPACE="ibm-system"

# create a tempdir
TEMP_DIR=$(mktemp -d)
trap cleanup EXIT

NODE=$1

shift
if [ "$1" == "--" ]; then
  shift
fi

COMMAND=$@

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
      nodeSelector:
        kubernetes.io/hostname: ${NODE}
      containers:
        - name: runon
          image: "debian:jessie"
          command:
            - sleep
            - 8h
          securityContext:
            privileged: true
      hostPID: true
      restartPolicy: Never
EOF
) > "${TEMP_DIR}"/job.yml

if ! kubectl apply -f "${TEMP_DIR}"/job.yml 2>&1 > /dev/null; then
  echo "unable to create job, bailing out"
  exit 1
fi


# get the uid
ID=$(kubectl get job "${JOB_NAME}" -n ${NAMESPACE} -o 'jsonpath={.metadata.uid}')
if [ -z "${ID}" ]; then
  echo "ERR unable to get job id"
  exit 1
fi


POD=$(kubectl get pods -n ${NAMESPACE} -l controller-uid="${ID}",job-name="${JOB_NAME}" -o 'jsonpath={.items[].metadata.name}')

SUCCESS=
current_time=$(date +%s)
stop_time=$((current_time + 60)) # this shouldn't take that long, give it 60 seconds

status=$(pod_status "${POD}")
while [[ $current_time -lt $stop_time ]]; do
  if [ "${status}" = "True" ]; then
    SUCCESS=true
    break
  else
    sleep 1
  fi
  status=$(pod_status "${POD}")
done

if [ "${SUCCESS}" = "true" ]; then
  kubectl exec -it "${POD}" -n ${NAMESPACE} -- nsenter -t 1 -m -u -i -n -p -- $@ <&0 2>&1
else
  echo "failed pod wasn't ready in time"
  exit 1
fi
