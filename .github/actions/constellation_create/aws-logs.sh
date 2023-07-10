#!/usr/bin/env bash

# Usage: ./aws-logs.sh <region>

set -euo pipefail
shopt -s inherit_errexit

echo "Using AWS region: ${1}"

# TODO(msanft): Remove once 2.9.0 is released
CP_SELECTOR="module.instance_group_control_plane"
W_SELECTOR="module.instance_group_worker_nodes"
if [[ $(./constellation version) != *"2.8.0"* ]]; then
  echo "Constellation version is not 2.8.0, using updated ASG selectors"
  CP_SELECTOR='module.instance_group["control_plane_default"]'
  W_SELECTOR='module.instance_group["worker_default"]'
fi

pushd constellation-terraform
controlAutoscalingGroup=$(
  terraform show -json |
    jq --arg selector "$CP_SELECTOR" \
      -r .'values.root_module.child_modules[] |
      select(.address == $selector) |
      .resources[0].values.name'
)
workerAutoscalingGroup=$(
  terraform show -json |
    jq --arg selector "$W_SELECTOR" \
      -r .'values.root_module.child_modules[] |
      select(.address == $selector) |
      .resources[0].values.name'
)
popd

controlInstances=$(
  aws autoscaling describe-auto-scaling-groups \
    --region "${1}" \
    --no-paginate \
    --output json \
    --auto-scaling-group-names "${controlAutoscalingGroup}" |
    jq -r '.AutoScalingGroups[0].Instances[].InstanceId'
)
workerInstances=$(
  aws autoscaling describe-auto-scaling-groups \
    --region "${1}" \
    --no-paginate \
    --output json \
    --auto-scaling-group-names "${workerAutoscalingGroup}" |
    jq -r '.AutoScalingGroups[0].Instances[].InstanceId'
)

echo "Fetching logs from control planes: ${controlInstances}"

for instance in ${controlInstances}; do
  printf "Fetching for %s\n" "${instance}"
  aws ec2 get-console-output --region "${1}" --instance-id "${instance}" |
    jq -r .'Output' |
    tail -n +2 > control-plane-"${instance}".log
done

echo "Fetching logs from worker nodes: ${workerInstances}"

for instance in ${workerInstances}; do
  printf "Fetching for %s\n" "${instance}"
  aws ec2 get-console-output --region "${1}" --instance-id "${instance}" |
    jq -r .'Output' |
    tail -n +2 > worker-"${instance}".log
done
