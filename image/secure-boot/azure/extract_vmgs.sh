#!/usr/bin/env bash
set -euo pipefail

if [ -z "${CONFIG_FILE-}" ] && [ -f "${CONFIG_FILE-}" ]; then
    . "${CONFIG_FILE}"
fi
AZURE_SUBSCRIPTION=$(az account show --query id -o tsv)
POSITIONAL_ARGS=()

while [[ $# -gt 0 ]]; do
  case $1 in
    -n|--name)
      AZURE_VM_NAME="$2"
      shift # past argument
      shift # past value
      ;;
    -*|--*)
      echo "Unknown option $1"
      exit 1
      ;;
    *)
      POSITIONAL_ARGS+=("$1") # save positional arg
      shift # past argument
      ;;
  esac
done

set -- "${POSITIONAL_ARGS[@]}" # restore positional parameters

VM_DISK=$(az vm show -g "${AZURE_RESOURCE_GROUP_NAME}" --name "${AZURE_VM_NAME}" --query "storageProfile.osDisk.managedDisk.id" -o tsv)
LOCATION=$(az disk show --ids "${VM_DISK}" --query "location" -o tsv)

az snapshot create \
    -g "${AZURE_RESOURCE_GROUP_NAME}" \
    --source "${VM_DISK}" \
    --name "${AZURE_SNAPSHOT_NAME}" \
    -l "${LOCATION}"

# Azure CLI does not implement getSecureVMGuestStateSAS for snapshots yet
# az snapshot grant-access \
#     --duration-in-seconds 3600 \
#     --access-level Read \
#     --name "${AZURE_SNAPSHOT_NAME}" \
#     -g "${AZURE_RESOURCE_GROUP_NAME}"

BEGIN=$(az rest \
    --method post \
    --url "https://management.azure.com/subscriptions/${AZURE_SUBSCRIPTION}/resourceGroups/${AZURE_RESOURCE_GROUP_NAME}/providers/Microsoft.Compute/snapshots/${AZURE_SNAPSHOT_NAME}/beginGetAccess" \
    --uri-parameters api-version="2021-12-01" \
    --body '{"access": "Read", "durationInSeconds": 3600, "getSecureVMGuestStateSAS": true}' \
    --verbose 2>&1)
ASYNC_OPERATION_URI=$(echo "${BEGIN}" | grep Azure-AsyncOperation | cut -d ' ' -f 7 | tr -d "'")
sleep 10
ACCESS=$(az rest --method get --url "${ASYNC_OPERATION_URI}")
VMGS_URL=$(echo "${ACCESS}" | jq -r '.properties.output.securityDataAccessSAS')

curl -L -o "${AZURE_VMGS_FILENAME}" "${VMGS_URL}"

az snapshot revoke-access \
    --name "${AZURE_SNAPSHOT_NAME}" \
    -g "${AZURE_RESOURCE_GROUP_NAME}"
az snapshot delete \
    --name "${AZURE_SNAPSHOT_NAME}" \
    -g "${AZURE_RESOURCE_GROUP_NAME}"
echo "VMGS saved to ${AZURE_VMGS_FILENAME}"
