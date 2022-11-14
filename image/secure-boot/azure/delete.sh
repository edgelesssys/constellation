#!/usr/bin/env bash

set -euo pipefail
shopt -s inherit_errexit

if [[ -z ${CONFIG_FILE-} ]] && [[ -f ${CONFIG_FILE-} ]]; then
  # shellcheck source=/dev/null
  . "${CONFIG_FILE}"
fi
POSITIONAL_ARGS=()

while [[ $# -gt 0 ]]; do
  case $1 in
  -n | --name)
    AZURE_VM_NAME="$2"
    shift # past argument
    shift # past value
    ;;
  -*)
    echo "Unknown option $1"
    exit 1
    ;;
  *)
    POSITIONAL_ARGS+=("$1") # save positional arg
    shift                   # past argument
    ;;
  esac
done

set -- "${POSITIONAL_ARGS[@]}" # restore positional parameters

AZ_VM_INFO=$(az vm show --name "${AZURE_VM_NAME}" --resource-group "${AZURE_RESOURCE_GROUP_NAME}" -o json)
NIC=$(echo "${AZ_VM_INFO}" | jq -r '.networkProfile.networkInterfaces[0].id')
NIC_INFO=$(az network nic show --ids "${NIC}" -o json)
PUBIP=$(echo "${NIC_INFO}" | jq -r '.ipConfigurations[0].publicIpAddress.id')
NSG=$(echo "${NIC_INFO}" | jq -r '.networkSecurityGroup.id')
SUBNET=$(echo "${NIC_INFO}" | jq -r '.ipConfigurations[0].subnet.id')
VNET=${SUBNET//\/subnets\/.*/}
DISK=$(echo "${AZ_VM_INFO}" | jq -r '.storageProfile.osDisk.managedDisk.id')

delete_vm() {
  az vm delete -y --name "${AZURE_VM_NAME}" \
    --resource-group "${AZURE_RESOURCE_GROUP_NAME}" || true
}

delete_vnet() {
  az network vnet delete --ids "${VNET}" || true
}

delete_subnet() {
  az network vnet subnet delete --ids "${SUBNET}" || true
}

delete_nsg() {
  az network nsg delete --ids "${NSG}" || true
}

delete_pubip() {
  az network public-ip delete --ids "${PUBIP}" || true
}

delete_disk() {
  az disk delete -y --ids "${DISK}" || true
}

delete_nic() {
  az network nic delete --ids "${NIC}" || true
}

delete_vm
delete_disk
delete_nic
delete_nsg
delete_subnet
delete_vnet
delete_pubip
