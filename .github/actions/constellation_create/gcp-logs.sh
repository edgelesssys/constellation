#!/usr/bin/env bash

CONTROL_INSTANCE_GROUP=$(terraform show -json | jq -r .'values.root_module.child_modules[] | select(.address == "module.instance_group_control_plane") | .resources[0].values.base_instance_name' )
WORKER_INSTANCE_GROUP=$(terraform show -json | jq -r .'values.root_module.child_modules[] | select(.address == "module.instance_group_worker") | .resources[0].values.base_instance_name')
ZONE=$(terraform show -json | jq -r .'values.root_module.child_modules[] | select(.address == "module.instance_group_control_plane") | .resources[0].values.zone' )

CONTROL_INSTANCE_GROUP_SHORT=${CONTROL_INSTANCE_GROUP##*/}
WORKER_INSTANCE_GROUP_SHORT=${WORKER_INSTANCE_GROUP##*/}

CONTROL_INSTANCES=$(gcloud compute instance-groups managed list-instances ${CONTROL_INSTANCE_GROUP_SHORT} --zone ${ZONE} --format=json | jq -r '.[] | .instance')
WORKER_INSTANCES=$(gcloud compute instance-groups managed list-instances ${WORKER_INSTANCE_GROUP_SHORT} --zone ${ZONE} --format=json | jq -r '.[] | .instance')

ALL_INSTANCES="$CONTROL_INSTANCES $WORKER_INSTANCES"

printf "Fetching logs for %s and %s\n" ${CONTROL_INSTANCES} ${WORKER_INSTANCES}

for INSTANCE in $ALL_INSTANCES; do
  SHORT_NAME=${INSTANCE##*/}
  printf "Fetching for %s\n" ${SHORT_NAME}
  gcloud compute instances get-serial-port-output ${INSTANCE} \
    --port 1 \
    --start 0 \
    --zone ${ZONE} > ${SHORT_NAME}.log
done
