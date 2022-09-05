#!/bin/bash

# Usage: ./gcp-logs.sh <control-instance-group-name> <worker-instance-group-name> <zone>

CONTROL_INSTANCES=$(gcloud compute instance-groups list-instances $1 --zone $3 --format=json | jq -r .'[] | .instance')
WORKER_INSTANCES=$(gcloud compute instance-groups list-instances $2 --zone $3 --format=json | jq -r .'[] | .instance')
ALL_INSTANCES="$CONTROL_INSTANCES $WORKER_INSTANCES"

printf "Fetching logs for %s and %s\n" $CONTROL_INSTANCES $WORKER_INSTANCES

for INSTANCE in $ALL_INSTANCES; do
  SHORT_NAME=${INSTANCE##*/}
  printf "Fetching for %s\n" $SHORT_NAME
  gcloud compute instances get-serial-port-output $INSTANCE \
    --port 1 \
    --start 0 \
    --zone $3 > $SHORT_NAME.log
done
