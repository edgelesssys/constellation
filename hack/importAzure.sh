#!/bin/bash

# importAzure imports a downloaded Azure VM image into Azure cloud.
# Parameters are provided via environment variables.
#
# Usage:
#  $ AZURE_IMAGE_VERSION=0.1.0 AZURE_RESOURCE_GROUP_NAME=constellation-images ./importAzure.sh
# Required values.
# * AZURE_RESOURCE_GROUP_NAME: (required) resource group in Azure to use. Needs to exist!
# * AZURE_IMAGE_VERSION: (required) version number used for uploaded image. <major>.<minor>.<patch>
# Optional values.
# * AZURE_IMAGE_FILE: (optional, default: ./abcd) Path to image file to be uploaded.
# * AZURE_REGION: (optional, default: westus) Region used in Azure.
# * AZURE_GALLERY_NAME: (optional, default: constellation_import) Name for Azure shared image gallery. Will be created as part of this script.
# * AZURE_IMAGE_NAME: (optional, default: upload-target) Temporary image used for upload, must not exist.

set -euo pipefail

# Required tools
if ! command -v az &> /dev/null
then
    echo "az CLI could not be found"
    echo "Please instal it from: https://docs.microsoft.com/en-us/cli/azure/install-azure-cli"
    exit
fi
if ! command -v azcopy &> /dev/null
then
    echo "azcopy could not be found"
    echo "Please instal it from: https://docs.microsoft.com/en-us/azure/storage/common/storage-use-azcopy-v10"
    exit
fi
if ! command -v jq &> /dev/null
then
    echo "jq could not be found"
    echo "Please instal it from: https://github.com/stedolan/jq"
    exit
fi

AZURE_IMAGE_FILE="${AZURE_IMAGE_FILE:-$(pwd)/abcd}"
AZURE_REGION="${AZURE_REGION:-westus}"
AZURE_GALLERY_NAME="${AZURE_GALLERY_NAME:-constellation_import}"
AZURE_PUBLISHER="${AZURE_PUBLISHER:-edgelesssys}"
AZURE_IMAGE_NAME="${AZURE_IMAGE_NAME:-upload-target}"
AZURE_IMAGE_OFFER="${AZURE_IMAGE_OFFER:-constellation}"
AZURE_IMAGE_DEFINITION="${AZURE_IMAGE_DEFINITION:-constellation}"
AZURE_SKU="${AZURE_SKU:-constellation-coreos}"
AZURE_SECURITY_TYPE="${AZURE_SECURITY_TYPE:-TrustedLaunch}"

if [[ -z "${AZURE_RESOURCE_GROUP_NAME}" ]]; then
  echo "Please provide a value for AZURE_RESOURCE_GROUP_NAME."
  exit 1
fi

if [[ -z "${AZURE_IMAGE_VERSION}" ]]; then
  echo "Please provide a value for AZURE_IMAGE_VERSION of pattern <major>.<minor>.<patch>"
  exit 1
fi


echo "Using following settings:"
echo "AZURE_REGION=${AZURE_REGION}"
echo "AZURE_RESOURCE_GROUP_NAME=${AZURE_RESOURCE_GROUP_NAME}"
echo "AZURE_GALLERY_NAME=${AZURE_GALLERY_NAME}"
echo "AZURE_IMAGE_FILE=${AZURE_IMAGE_FILE}"
echo "AZURE_IMAGE_NAME=${AZURE_IMAGE_NAME}"
echo "AZURE_IMAGE_OFFER=${AZURE_IMAGE_OFFER}"
echo "AZURE_IMAGE_DEFINITION=${AZURE_IMAGE_DEFINITION}"
echo "AZURE_IMAGE_VERSION=${AZURE_IMAGE_VERSION}"
echo "AZURE_PUBLISHER=${AZURE_PUBLISHER}"
echo "AZURE_SKU=${AZURE_SKU}"
echo "AZURE_SECURITY_TYPE=${AZURE_SECURITY_TYPE}"
echo ""

read -p "Continue (y/n)?" choice
case "$choice" in
  y|Y ) echo "Starting import...";;
  n|N ) echo "Abort!"; exit 1;;
  * ) echo "invalid"; exit 1;;
esac

echo "Preparing to upload '${AZURE_IMAGE_FILE} to Azure."

SIZE=$(wc -c ${AZURE_IMAGE_FILE} | cut -d " " -f1)
echo "Size is ${SIZE} bytes."

echo "Creating disk (${AZURE_IMAGE_NAME}) as import target."
az disk create -n ${AZURE_IMAGE_NAME} -g ${AZURE_RESOURCE_GROUP_NAME} -l ${AZURE_REGION} --hyper-v-generation V2 --os-type Linux --for-upload --upload-size-bytes ${SIZE} --sku standard_lrs
echo "Waiting for disk to be created."
az disk wait --created -n ${AZURE_IMAGE_NAME} -g ${AZURE_RESOURCE_GROUP_NAME}
echo "Retrieving disk ID."
AZURE_DISK_ID=$(az disk list --query "[?name == '${AZURE_IMAGE_NAME}' && resourceGroup == '${AZURE_RESOURCE_GROUP_NAME^^}'] | [0].id" --output json | jq -r)
echo "Disk ID is ${AZURE_DISK_ID}"

echo "Generating SAS URL for authorized upload."
AZURE_SAS_URL=$(az disk grant-access -n ${AZURE_IMAGE_NAME} -g ${AZURE_RESOURCE_GROUP_NAME} --access-level Write --duration-in-seconds 86400 | jq -r .accessSas)
echo "Uploading image file to Azure disk."
azcopy copy ${AZURE_IMAGE_FILE} ${AZURE_SAS_URL} --blob-type PageBlob
echo "Finalizing upload."
az disk revoke-access -n ${AZURE_IMAGE_NAME} -g ${AZURE_RESOURCE_GROUP_NAME}

echo "Creating Azure image."
az image create -g ${AZURE_RESOURCE_GROUP_NAME} -l ${AZURE_REGION} -n ${AZURE_IMAGE_NAME} --hyper-v-generation V2 --os-type Linux --source ${AZURE_DISK_ID}
echo "Creating Azure Shared Image Gallery."
az sig create -l ${AZURE_REGION} --gallery-name ${AZURE_GALLERY_NAME} --resource-group ${AZURE_RESOURCE_GROUP_NAME}
echo "Creating Image Definition."
az sig image-definition create --resource-group ${AZURE_RESOURCE_GROUP_NAME} -l ${AZURE_REGION} --gallery-name ${AZURE_GALLERY_NAME} --gallery-image-definition ${AZURE_IMAGE_DEFINITION} --publisher ${AZURE_PUBLISHER} --offer ${AZURE_IMAGE_OFFER} --sku ${AZURE_SKU} --os-type Linux --os-state generalized --hyper-v-generation V2 --features SecurityType=${AZURE_SECURITY_TYPE}
echo "Retrieving temporary image ID."
AZURE_IMAGE_ID=$(az image list --query "[?name == '${AZURE_IMAGE_NAME}' && resourceGroup == '${AZURE_RESOURCE_GROUP_NAME^^}'] | [0].id" --output json | jq -r)

echo "Creating final image version."
az sig image-version create --resource-group ${AZURE_RESOURCE_GROUP_NAME} -l ${AZURE_REGION} --gallery-name ${AZURE_GALLERY_NAME} --gallery-image-definition ${AZURE_IMAGE_DEFINITION} --gallery-image-version ${AZURE_IMAGE_VERSION} --target-regions ${AZURE_REGION} --replica-count 1 --managed-image ${AZURE_IMAGE_ID}

echo "Cleaning up ephemeral resources."
az image delete --ids ${AZURE_IMAGE_ID}
az disk delete -y --ids ${AZURE_DISK_ID}

IMAGE_VERSION=$(az sig image-version show --resource-group ${AZURE_RESOURCE_GROUP_NAME}  --gallery-name ${AZURE_GALLERY_NAME} --gallery-image-definition ${AZURE_IMAGE_DEFINITION} --gallery-image-version ${AZURE_IMAGE_VERSION} -o tsv --query id)
echo "Image ID is ${IMAGE_VERSION}"

# # Cleanup all
# az sig image-version delete --resource-group ${AZURE_RESOURCE_GROUP_NAME} --gallery-image-definition ${AZURE_IMAGE_DEFINITION} --gallery-image-version ${AZURE_IMAGE_VERSION} --gallery-name ${AZURE_GALLERY_NAME}
# az sig image-definition delete --resource-group ${AZURE_RESOURCE_GROUP_NAME} --gallery-name ${AZURE_GALLERY_NAME} --gallery-image-definition ${AZURE_IMAGE_DEFINITION}
# az sig delete --resource-group ${AZURE_RESOURCE_GROUP_NAME} --gallery-name ${AZURE_GALLERY_NAME}
