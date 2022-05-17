#!/bin/bash

set -o xtrace
trap 'terminate $?' ERR

terminate() {
    echo "error: $1"
    constellation terminate
    exit 1
}

main() {
    command -v constellation > /dev/null
    command -v go > /dev/null
    command -v jq > /dev/null

    mkdir -p ./pcrs

    # Fetch Azure PCRs
    # TODO: Switch to confidential VMs
    constellation create azure 2 Standard_D4s_v3 --name pcr-fetch -y
    coord_ip=$(jq '.azurecoordinators | to_entries[] | select(.key|startswith("")) | .value.PublicIP' -rcM constellation-state.json)
    go run main.go -coord-ip "${coord_ip}" -o ./pcrs/azure_pcrs.go
    constellation terminate

    # Fetch GCP PCRs
    constellation create gcp 2 n2d-standard-2 --name pcr-fetch -y
    coord_ip=$(jq '.gcpcoordinators | to_entries[] | select(.key|startswith("")) | .value.PublicIP' -rcM constellation-state.json)
    go run main.go -coord-ip "${coord_ip}" -o ./pcrs/gcp_pcrs.go
    constellation terminate
}

main
