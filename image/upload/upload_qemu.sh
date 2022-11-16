#!/usr/bin/env bash
# Copyright (c) Edgeless Systems GmbH
#
# SPDX-License-Identifier: AGPL-3.0-only

set -euo pipefail
shopt -s inherit_errexit

if [[ -z ${CONFIG_FILE-} ]] && [[ -f ${CONFIG_FILE-} ]]; then
  # shellcheck source=/dev/null
  . "${CONFIG_FILE}"
fi

aws s3 cp "${QEMU_IMAGE_PATH}" "s3://${QEMU_BUCKET}/v1/raw/${IMAGE_VERSION_UID}/qemu/image.raw" --no-progress

image_url="${QEMU_BASE_URL}/v1/raw/${IMAGE_VERSION_UID}/qemu/image.raw"

json=$(jq -ncS \
  --arg image_url "${image_url}" \
  '{"qemu": {"default": $image_url}}')
echo -n "${json}" > "${QEMU_JSON_OUTPUT}"
