#!/usr/bin/env bash

set -euo pipefail
shopt -s inherit_errexit

pushd constellation-terraform
controlInstanceGroup=$(
  terraform show -json |
    jq -r .'values.root_module.child_modules[] |
    select(.address == "module.instance_group_control_plane") |
    .resources[0].values.base_instance_name'
)
workerInstanceGroup=$(
  terraform show -json |
    jq -r .'values.root_module.child_modules[] |
    select(.address == "module.instance_group_worker") |
     .resources[0].values.base_instance_name'
)
zone=$(
  terraform show -json |
    jq -r .'values.root_module.child_modules[] |
    select(.address == "module.instance_group_control_plane") |
     .resources[0].values.zone'
)
popd

controlInstances=$(
  gcloud compute instance-groups managed list-instances "${controlInstanceGroup##*/}" \
    --zone "${zone}" \
    --format=json |
    jq -r '.[] | .instance'
)
workerInstances=$(
  gcloud compute instance-groups managed list-instances "${workerInstanceGroup##*/}" \
    --zone "${zone}" \
    --format=json |
    jq -r '.[] | .instance'
)

allInstances="${controlInstances} ${workerInstances}"

printf "Fetching logs for %s and %s\n" "${controlInstances}" "${workerInstances}"

for instance in ${allInstances}; do
  shortName=${instance##*/}
  printf "Fetching for %s\n" "${shortName}"
  gcloud compute instances get-serial-port-output "${instance}" \
    --port 1 \
    --start 0 \
    --zone "${zone}" > "${shortName}".log
done
