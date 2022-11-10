#!/usr/bin/env bash

# Usage: ./aws-logs.sh <region>

controlAutoscalingGroup=$(\
    terraform show -json | \
    jq -r .'values.root_module.child_modules[] |
        select(.address == "module.instance_group_control_plane") |
        .resources[0].values.name' \
)
workerAutoscalingGroup=$(\
    terraform show -json | \
    jq -r .'values.root_module.child_modules[] |
        select(.address == "module.instance_group_worker_nodes") |
        .resources[0].values.name' \
)

controlInstances=$(\
    aws autoscaling describe-auto-scaling-groups \
        --region "${1}" \
        --no-paginate \
        --output json \
        --auto-scaling-group-names "${controlAutoscalingGroup}" | \
    jq -r '.AutoScalingGroups[0].Instances[].InstanceId' \
)
workerInstances=$(\
    aws autoscaling describe-auto-scaling-groups \
        --region "${1}" \
        --no-paginate \
        --output json \
        --auto-scaling-group-names "${workerAutoscalingGroup}" | \
    jq -r '.AutoScalingGroups[0].Instances[].InstanceId' \
)

echo "Fetching logs from control planes: ${controlInstances}"

for instance in ${controlInstances}; do
    printf "Fetching for %s\n" "${instance}"
    aws ec2 get-console-output --region "${1}" --instance-id "${instance}" | \
        jq -r .'Output' | \
        tail -n +2 > control-plane-"${instance}".log
done

echo "Fetching logs from worker nodes: ${workerInstances}"

for instance in ${workerInstances}; do
    printf "Fetching for %s\n" "${instance}"
    aws ec2 get-console-output --region "${1}" --instance-id "${instance}" | \
        jq -r .'Output' | \
        tail -n +2 > worker-"${instance}".log
done
