#!/usr/bin/env bash
# Copyright (c) Edgeless Systems GmbH
#
# SPDX-License-Identifier: AGPL-3.0-only

set -euo pipefail
shopt -s inherit_errexit

if [[ -f ${CONFIG_FILE-} ]]; then
  # shellcheck source=/dev/null
  . "${CONFIG_FILE}"
fi

CONTAINERS_JSON=$(mktemp /tmp/containers-XXXXXXXXXXXXXX.json)
declare -A AMI_FOR_REGION

import_status() {
  local import_task_id=$1
  aws ec2 describe-import-snapshot-tasks --region "${AWS_REGION}" --import-task-ids "${import_task_id}" | jq -r '.ImportSnapshotTasks[0].SnapshotTaskDetail.Status'
}

wait_for_import() {
  local import_task_id=$1
  local status
  echo -n "Waiting for import to finish"
  while true; do
    local status
    status=$(import_status "${import_task_id}")
    case "${status}" in
    completed)
      echo -e "\nImport completed."
      break
      ;;
    active)
      echo -n "."
      sleep 5
      ;;
    *)
      echo "Unexpected status: ${status}"
      exit 1
      ;;
    esac
  done
}

wait_for_image_available() {
  local ami_id=$1
  local region=$2
  echo -n "Waiting for image ${ami_id} to be available"
  while true; do
    # Waiter ImageAvailable failed: Max attempts exceeded
    local status
    status=$(aws ec2 wait image-available \
      --region "${region}" \
      --image-ids "${ami_id}" 2>&1 || true)
    case "${status}" in
    "")
      echo -e "\nImage available."
      break
      ;;
    *"Max attempts exceeded"*)
      echo -n "."
      ;;
    *)
      echo "Unexpected status: ${status}"
      exit 1
      ;;
    esac
  done
}

tag_ami_with_backing_snapshot() {
  local ami_id=$1
  local region=$2
  wait_for_image_available "${ami_id}" "${region}"
  local snapshot_id
  snapshot_id=$(aws ec2 describe-images \
    --region "${region}" \
    --image-ids "${ami_id}" \
    --output text --query "Images[0].BlockDeviceMappings[0].Ebs.SnapshotId")
  aws ec2 create-tags \
    --region "${region}" \
    --resources "${ami_id}" "${snapshot_id}" \
    --tags "Key=Name,Value=${AWS_IMAGE_NAME}"
}

make_ami_public() {
  local ami_id=$1
  local region=$2
  if [[ ${AWS_PUBLISH-} != "true" ]]; then
    return
  fi
  aws ec2 modify-image-attribute \
    --region "${region}" \
    --image-id "${ami_id}" \
    --launch-permission "Add=[{Group=all}]"
}

create_ami_from_raw_disk() {
  echo "Uploading raw disk image to S3"
  aws s3 cp "${AWS_IMAGE_PATH}" "s3://${AWS_BUCKET}/${AWS_IMAGE_FILENAME}" --no-progress
  printf '{
        "Description": "%s",
        "Format": "raw",
        "UserBucket": {
            "S3Bucket": "%s",
            "S3Key": "%s"
        }
    }' "${AWS_IMAGE_NAME}" "${AWS_BUCKET}" "${AWS_IMAGE_FILENAME}" > "${CONTAINERS_JSON}"
  IMPORT_SNAPSHOT=$(aws ec2 import-snapshot --region "${AWS_REGION}" --disk-container "file://${CONTAINERS_JSON}")
  echo "${IMPORT_SNAPSHOT}"
  IMPORT_TASK_ID=$(echo "${IMPORT_SNAPSHOT}" | jq -r '.ImportTaskId')
  aws ec2 describe-import-snapshot-tasks --region "${AWS_REGION}" --import-task-ids "${IMPORT_TASK_ID}"
  wait_for_import "${IMPORT_TASK_ID}"
  AWS_SNAPSHOT=$(aws ec2 describe-import-snapshot-tasks --region "${AWS_REGION}" --import-task-ids "${IMPORT_TASK_ID}" | jq -r '.ImportSnapshotTasks[0].SnapshotTaskDetail.SnapshotId')
  echo "Deleting raw disk image from S3"
  aws s3 rm "s3://${AWS_BUCKET}/${AWS_IMAGE_FILENAME}"
  rm "${CONTAINERS_JSON}"
  REGISTER_OUT=$(
    aws ec2 register-image \
      --region "${AWS_REGION}" \
      --name "${AWS_IMAGE_NAME}" \
      --boot-mode uefi \
      --architecture x86_64 \
      --root-device-name /dev/xvda \
      --block-device-mappings "DeviceName=/dev/xvda,Ebs={SnapshotId=${AWS_SNAPSHOT}}" \
      --ena-support \
      --tpm-support v2.0 \
      --uefi-data "$(cat "${AWS_EFIVARS_PATH}")"
  )
  IMAGE_ID=$(echo "${REGISTER_OUT}" | jq -r '.ImageId')
  AMI_FOR_REGION=(["${AWS_REGION}"]="${IMAGE_ID}")
  tag_ami_with_backing_snapshot "${IMAGE_ID}" "${AWS_REGION}"
  make_ami_public "${IMAGE_ID}" "${AWS_REGION}"
  echo "Imported initial AMI as ${IMAGE_ID} in ${AWS_REGION}"
}

replicate_ami() {
  local target_region=$1
  local replicated_image_out
  replicated_image_out=$(aws ec2 copy-image \
    --name "${AWS_IMAGE_NAME}" \
    --source-region "${AWS_REGION}" \
    --source-image-id "${IMAGE_ID}" \
    --region "${target_region}")
  local replicated_image_id
  replicated_image_id=$(echo "${replicated_image_out}" | jq -r '.ImageId')
  AMI_FOR_REGION["${target_region}"]=${replicated_image_id}
  echo "Replicated AMI as ${replicated_image_id} in ${target_region}"
}

create_ami_from_raw_disk
# replicate in parallel
for region in ${AWS_REPLICATION_REGIONS}; do
  replicate_ami "${region}"
done
# wait for all images to be available and tag + publish them
for region in ${AWS_REPLICATION_REGIONS}; do
  tag_ami_with_backing_snapshot "${AMI_FOR_REGION[${region}]}" "${region}"
  make_ami_public "${AMI_FOR_REGION[${region}]}" "${region}"
done

json=$(jq -ncS \
  --arg region "${AWS_REGION}" \
  --arg ami "${AMI_FOR_REGION[${AWS_REGION}]}" \
  '{"aws":{($region): $ami}}')
for region in ${AWS_REPLICATION_REGIONS}; do
  json=$(jq -ncS \
    --argjson json "${json}" \
    --arg region "${region}" \
    --arg ami "${AMI_FOR_REGION[${region}]}" \
    '$json * {"aws": {($region): $ami}}')
done

echo "${json}" > "${AWS_JSON_OUTPUT}"
