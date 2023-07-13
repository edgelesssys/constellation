#!/usr/bin/env bash

# Usage: ./aws-logs.sh <region>

set -euo pipefail
shopt -s inherit_errexit

echo "Using AWS region: ${1}"
echo "Using Constellation UID: ${2}"

controlInstances=$(
  aws ec2 describe-instances \
    --filters "Name=tag:constellation-uid,Values=${2}" "Name=tag:constellation-role,Values=control-plane" \
    --region "${1}" \
    --no-paginate \
    --output json |
    jq -r '.Reservations[].Instances[].InstanceId'
)
workerInstances=$(
  aws ec2 describe-instances \
    --filters "Name=tag:constellation-uid,Values=${2}" "Name=tag:constellation-role,Values=worker" \
    --region "${1}" \
    --no-paginate \
    --output json |
    jq -r '.Reservations[].Instances[].InstanceId'
)

echo "Fetching logs from control planes"

for instance in ${controlInstances}; do
  printf "Fetching for %s\n" "${instance}"
  aws ec2 get-console-output --region "${1}" --instance-id "${instance}" |
    jq -r .'Output' |
    tail -n +2 > control-plane-"${instance}".log
done

echo "Fetching logs from worker nodes"

for instance in ${workerInstances}; do
  printf "Fetching for %s\n" "${instance}"
  aws ec2 get-console-output --region "${1}" --instance-id "${instance}" |
    jq -r .'Output' |
    tail -n +2 > worker-"${instance}".log
done
