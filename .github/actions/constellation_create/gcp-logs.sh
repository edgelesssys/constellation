#!/usr/bin/env bash

set -euo pipefail
shopt -s inherit_errexit

echo "Using Zone: ${1}"
echo "Using Constellation UID: ${2}"

allInstances=$(
  gcloud compute instances list \
    --filter="labels.constellation-uid=${2}" \
    --format=json | yq eval '.[] | .name' -
)

for instance in ${allInstances}; do
  printf "Fetching for %s\n" "${instance}"
  gcloud compute instances get-serial-port-output "${instance}" \
    --port 1 \
    --start 0 \
    --zone "${1}" > "${instance}".log
done
