#!/usr/bin/env bash

set -euo pipefail
shopt -s inherit_errexit

printf "Fetching logs of instances in resource group %s\n" "${1}"

# get list of all scale sets
scalesetsjson=$(az vmss list --resource-group "${1}" -o json)
scalesetslist=$(echo "${scalesetsjson}" | yq eval '.[] | .name' -)
subscription=$(az account show | yq eval .id -)

printf "Checking scalesets %s\n" "${scalesetslist}"

for scaleset in ${scalesetslist}; do
  instanceids=$(
    az vmss list-instances \
      --resource-group "${1}" \
      --name "${scaleset}" \
      -o json |
      yq eval '.[] | .instanceId' -
  )
  printf "Checking instance IDs %s\n" "${instanceids}"
  for instanceid in ${instanceids}; do
    bloburi=$(
      az rest \
        --method post \
        --url https://management.azure.com/subscriptions/"${subscription}"/resourceGroups/"${1}"/providers/Microsoft.Compute/virtualMachineScaleSets/"${scaleset}"/virtualmachines/"${instanceid}"/retrieveBootDiagnosticsData?api-version=2022-03-01 |
        yq eval '.serialConsoleLogBlobUri' -
    )
    sleep 4
    curl -fsSL -o "./${scaleset}-${instanceid}.log" "${bloburi}"
    realpath "./${scaleset}-${instanceid}.log"
  done
done
