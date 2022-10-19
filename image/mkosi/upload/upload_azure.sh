#!/usr/bin/env bash
set -euo pipefail

if [ -z "${CONFIG_FILE-}" ] && [ -f "${CONFIG_FILE-}" ]; then
    . "${CONFIG_FILE}"
fi


CREATE_SIG_VERSION=NO
POSITIONAL_ARGS=()

while [[ $# -gt 0 ]]; do
  case $1 in
    -g|--gallery)
      CREATE_SIG_VERSION=YES
      shift # past argument
      ;;
    --disk-name)
      AZURE_DISK_NAME="$2"
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

if [[ "${AZURE_SECURITY_TYPE}" == "ConfidentialVM" ]]; then
    AZURE_DISK_SECURITY_TYPE=ConfidentialVM_VMGuestStateOnlyEncryptedWithPlatformKey
    AZURE_SIG_VERSION_ENCRYPTION_TYPE=EncryptedVMGuestStateOnlyWithPmk
elif [[ "${AZURE_SECURITY_TYPE}" == "ConfidentialVMSupported" ]]; then
    AZURE_DISK_SECURITY_TYPE=""
elif [[ "${AZURE_SECURITY_TYPE}" == "TrustedLaunch" ]]; then
    AZURE_DISK_SECURITY_TYPE=TrustedLaunch
else
    echo "Unknown security type: ${AZURE_SECURITY_TYPE}"
    exit 1
fi

AZURE_CVM_ENCRYPTION_ARGS=""
if [[ -n "${AZURE_SIG_VERSION_ENCRYPTION_TYPE-}" ]]; then
    AZURE_CVM_ENCRYPTION_ARGS=" --target-region-cvm-encryption "
    for region in ${AZURE_REPLICATION_REGIONS}; do
        AZURE_CVM_ENCRYPTION_ARGS=" ${AZURE_CVM_ENCRYPTION_ARGS} ${AZURE_SIG_VERSION_ENCRYPTION_TYPE}, "
    done
fi
echo "Replicating image in ${AZURE_REPLICATION_REGIONS}"

AZURE_VMGS_PATH=$1
if [[ -z "${AZURE_VMGS_PATH}" ]] && [[ "${AZURE_SECURITY_TYPE}" == "ConfidentialVM"  ]]; then
    echo "No VMGS path provided - using default ConfidentialVM VMGS"
    AZURE_VMGS_PATH="${BLOBS_DIR}/cvm-vmgs.vhd"
elif [[ -z "${AZURE_VMGS_PATH}" ]] && [[ "${AZURE_SECURITY_TYPE}" == "TrustedLaunch"  ]]; then
    echo "No VMGS path provided - using default TrsutedLaunch VMGS"
    AZURE_VMGS_PATH="${BLOBS_DIR}/trusted-launch-vmgs.vhd"
fi

SIZE=$(wc -c "${AZURE_IMAGE_PATH}" | cut -d " " -f1)

create_disk_with_vmgs () {
    az disk create \
        -n "${AZURE_DISK_NAME}" \
        -g "${AZURE_RESOURCE_GROUP_NAME}" \
        -l "${AZURE_REGION}" \
        --hyper-v-generation V2 \
        --os-type Linux \
        --upload-size-bytes "${SIZE}" \
        --sku standard_lrs \
        --upload-type UploadWithSecurityData \
        --security-type "${AZURE_DISK_SECURITY_TYPE}"
    az disk wait --created -n "${AZURE_DISK_NAME}" -g "${AZURE_RESOURCE_GROUP_NAME}"
    az disk list --output table --query "[?name == '${AZURE_DISK_NAME}' && resourceGroup == '${AZURE_RESOURCE_GROUP_NAME^^}']"
    DISK_SAS=$(az disk grant-access -n ${AZURE_DISK_NAME} -g ${AZURE_RESOURCE_GROUP_NAME} \
        --access-level Write --duration-in-seconds 86400 \
         ${AZURE_VMGS_PATH+"--secure-vm-guest-state-sas"})
    azcopy copy "${AZURE_IMAGE_PATH}" \
        "$(echo $DISK_SAS | jq -r .accessSas)" \
        --blob-type PageBlob
    if [[ -z "${AZURE_VMGS_PATH}" ]]; then
        echo "No VMGS path provided - skipping VMGS upload"
    else
        azcopy copy "${AZURE_VMGS_PATH}" \
            "$(echo $DISK_SAS | jq -r .securityDataAccessSas)" \
            --blob-type PageBlob
    fi
    az disk revoke-access -n "${AZURE_DISK_NAME}" -g "${AZURE_RESOURCE_GROUP_NAME}"
}

create_disk_without_vmgs () {
    az disk create \
        -n "${AZURE_DISK_NAME}" \
        -g "${AZURE_RESOURCE_GROUP_NAME}" \
        -l "${AZURE_REGION}" \
        --hyper-v-generation V2 \
        --os-type Linux \
        --upload-size-bytes "${SIZE}" \
        --sku standard_lrs \
        --upload-type Upload
    az disk wait --created -n "${AZURE_DISK_NAME}" -g "${AZURE_RESOURCE_GROUP_NAME}"
    az disk list --output table --query "[?name == '${AZURE_DISK_NAME}' && resourceGroup == '${AZURE_RESOURCE_GROUP_NAME^^}']"
    DISK_SAS=$(az disk grant-access -n ${AZURE_DISK_NAME} -g ${AZURE_RESOURCE_GROUP_NAME} \
        --access-level Write --duration-in-seconds 86400)
    azcopy copy "${AZURE_IMAGE_PATH}" \
        "$(echo $DISK_SAS | jq -r .accessSas)" \
        --blob-type PageBlob
    az disk revoke-access -n "${AZURE_DISK_NAME}" -g "${AZURE_RESOURCE_GROUP_NAME}"
}

create_disk () {
    if [[ -z "${AZURE_VMGS_PATH}" ]]; then
        create_disk_without_vmgs
    else
        create_disk_with_vmgs
    fi
}

delete_disk () {
    az disk delete -y -n "${AZURE_DISK_NAME}" -g "${AZURE_RESOURCE_GROUP_NAME}"
}

create_image () {
    if [[ -n "${AZURE_VMGS_PATH}" ]]; then
        return
    fi
    az image create \
        --resource-group ${AZURE_RESOURCE_GROUP_NAME} \
        -l ${AZURE_REGION} \
        -n ${AZURE_DISK_NAME} \
        --hyper-v-generation V2 \
        --os-type Linux \
        --source "$(az disk list --query "[?name == '${AZURE_DISK_NAME}' && resourceGroup == '${AZURE_RESOURCE_GROUP_NAME^^}'] | [0].id" --output tsv)"
}

delete_image () {
    if [[ -n "${AZURE_VMGS_PATH}" ]]; then
        return
    fi
    az image delete -n "${AZURE_DISK_NAME}" -g "${AZURE_RESOURCE_GROUP_NAME}"
}

create_sig_version () {
    if [[ -n "${AZURE_VMGS_PATH}" ]]; then
        local DISK="$(az disk list --query "[?name == '${AZURE_DISK_NAME}' && resourceGroup == '${AZURE_RESOURCE_GROUP_NAME^^}'] | [0].id" --output tsv)"
        local SOURCE="--os-snapshot ${DISK}"
    else
        local IMAGE="$(az image list --query "[?name == '${AZURE_DISK_NAME}' && resourceGroup == '${AZURE_RESOURCE_GROUP_NAME^^}'] | [0].id" --output tsv)"
        local SOURCE="--managed-image ${IMAGE}"
    fi
    az sig create -l "${AZURE_REGION}" --gallery-name "${AZURE_GALLERY_NAME}" --resource-group "${AZURE_RESOURCE_GROUP_NAME}" || true
    az sig image-definition create \
        --resource-group "${AZURE_RESOURCE_GROUP_NAME}" \
        -l "${AZURE_REGION}" \
        --gallery-name "${AZURE_GALLERY_NAME}" \
        --gallery-image-definition "${AZURE_IMAGE_DEFINITION}" \
        --publisher "${AZURE_PUBLISHER}" \
        --offer "${AZURE_IMAGE_OFFER}" \
        --sku "${AZURE_SKU}" \
        --os-type Linux \
        --os-state generalized \
        --hyper-v-generation V2 \
        --features SecurityType="${AZURE_SECURITY_TYPE}" || true
    az sig image-version create \
        --resource-group "${AZURE_RESOURCE_GROUP_NAME}" \
        -l "${AZURE_REGION}" \
        --gallery-name "${AZURE_GALLERY_NAME}" \
        --gallery-image-definition "${AZURE_IMAGE_DEFINITION}" \
        --gallery-image-version "${AZURE_IMAGE_VERSION}" \
        --target-regions ${AZURE_REPLICATION_REGIONS} \
        ${AZURE_CVM_ENCRYPTION_ARGS} \
        --replica-count 1 \
        --replication-mode Full \
        ${SOURCE}
}

create_disk

if [ "$CREATE_SIG_VERSION" = "YES" ]; then
    create_image
    create_sig_version
    delete_image
    delete_disk
fi
