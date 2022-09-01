#!/bin/bash

# get list of all scale sets
scalesetsjson=$(az vmss list --resource-group $1 -o json)
scalesetslist=$(echo $scalesetsjson | jq -r '.[] | .name')
subscription=$(az account  show | jq -r .id)

for scaleset in $scalesetslist; do
    instanceids=$(az vmss list-instances --resource-group $1 --name ${scaleset} -o json | jq -r '.[] | .instanceId')
    for instanceid in $instanceids; do
        bloburi=$(az rest --method post --url https://management.azure.com/subscriptions/${subscription}/resourceGroups/${1}/providers/Microsoft.Compute/virtualMachineScaleSets/${scaleset}/virtualmachines/$instanceid/retrieveBootDiagnosticsData?api-version=2022-03-01 | jq '.serialConsoleLogBlobUri' -r)
        sleep 4
        curl -sL -o "./${scaleset}-${instanceid}.log" $bloburi
        echo $(realpath "./${scaleset}-${instanceid}.log")
    done
done
