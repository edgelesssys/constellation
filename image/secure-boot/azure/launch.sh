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
  -g | --gallery)
    CREATE_FROM_GALLERY=YES
    shift # past argument
    ;;
  -d | --disk)
    CREATE_FROM_GALLERY=NO
    shift # past argument
    ;;
  --secure-boot)
    AZURE_SECURE_BOOT="$2"
    shift # past argument
    shift # past value
    ;;
  --disk-name)
    AZURE_DISK_NAME="$2"
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

if [[ ${AZURE_SECURITY_TYPE} == "ConfidentialVM" ]]; then
  VMSIZE="Standard_DC2as_v5"
else
  echo "Unknown security type: ${AZURE_SECURITY_TYPE}"
  exit 1
fi

create_vm_from_disk() {
  AZURE_DISK_REFERENCE=$(az disk show --resource-group "${AZURE_RESOURCE_GROUP_NAME}" --name "${AZURE_DISK_NAME}" --query id -o tsv)
  az vm create --name "${AZURE_VM_NAME}" \
    --resource-group "${AZURE_RESOURCE_GROUP_NAME}" \
    -l "${AZURE_REGION}" \
    --size "${VMSIZE}" \
    --public-ip-sku Standard \
    --os-type Linux \
    --attach-os-disk "${AZURE_DISK_REFERENCE}" \
    --security-type "${AZURE_SECURITY_TYPE}" \
    --os-disk-security-encryption-type VMGuestStateOnly \
    --enable-vtpm true \
    --enable-secure-boot "${AZURE_SECURE_BOOT}" \
    --boot-diagnostics-storage "" \
    --no-wait
}

create_vm_from_sig() {
  AZURE_IMAGE_REFERENCE=$(az sig image-version show \
    --gallery-image-definition "${AZURE_IMAGE_DEFINITION}" \
    --gallery-image-version "${AZURE_IMAGE_VERSION}" \
    --gallery-name "${AZURE_GALLERY_NAME}" \
    -g "${AZURE_RESOURCE_GROUP_NAME}" \
    --query id -o tsv)
  az vm create --name "${AZURE_VM_NAME}" \
    --resource-group "${AZURE_RESOURCE_GROUP_NAME}" \
    -l "${AZURE_REGION}" \
    --size "${VMSIZE}" \
    --public-ip-sku Standard \
    --image "${AZURE_IMAGE_REFERENCE}" \
    --security-type "${AZURE_SECURITY_TYPE}" \
    --os-disk-security-encryption-type VMGuestStateOnly \
    --enable-vtpm true \
    --enable-secure-boot "${AZURE_SECURE_BOOT}" \
    --boot-diagnostics-storage "" \
    --no-wait
}

if [[ ${CREATE_FROM_GALLERY} == "YES" ]]; then
  create_vm_from_sig
else
  create_vm_from_disk
fi

sleep 30
az vm boot-diagnostics enable --name "${AZURE_VM_NAME}" --resource-group "${AZURE_RESOURCE_GROUP_NAME}"
