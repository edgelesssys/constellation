#!/usr/bin/env bash

set -euo pipefail
shopt -s inherit_errexit

controlInstanceGroup=$(terraform show -json | jq -r .'values.root_module.child_modules[] | select(.address == "module.instance_group_control_plane") | .resources[0].values.base_instance_name' )
workerInstanceGroup=$(terraform show -json | jq -r .'values.root_module.child_modules[] | select(.address == "module.instance_group_worker") | .resources[0].values.base_instance_name')
zone=$(terraform show -json | jq -r .'values.root_module.child_modules[] | select(.address == "module.instance_group_control_plane") | .resources[0].values.zone' )

controlInstanceGroup=${controlInstanceGroup##*/}
workerInstanceGroupShort=${workerInstanceGroup##*/}

controlInstances=$(gcloud compute instance-groups managed list-instances "${controlInstanceGroup}" --zone "${zone}" --format=json | jq -r '.[] | .instance')
workerInstances=$(gcloud compute instance-groups managed list-instances "${workerInstanceGroupShort}" --zone "${zone}" --format=json | jq -r '.[] | .instance')

ALL_INSTANCES="${controlInstances} ${workerInstances}"

printf "Fetching logs for %s and %s\n" "${controlInstances}" "${workerInstances}"

for INSTANCE in ${ALL_INSTANCES}; do
  SHORT_NAME=${INSTANCE##*/}
  printf "Fetching for %s\n" "${SHORT_NAME}"
  gcloud compute instances get-serial-port-output "${INSTANCE}" \
    --port 1 \
    --start 0 \
    --zone "${zone}" > "${SHORT_NAME}".log
done
