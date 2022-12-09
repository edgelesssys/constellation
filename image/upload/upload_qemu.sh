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

path="constellation/v1/ref/${REF}/stream/${STREAM}/image/${IMAGE_VERSION}/csp/qemu/image.raw"
aws s3 cp "${QEMU_IMAGE_PATH}" "s3://${QEMU_BUCKET}/${path}" --no-progress

image_url="${QEMU_BASE_URL}/${path}"

json=$(jq -ncS \
  --arg image_url "${image_url}" \
  '{"qemu": {"default": $image_url}}')
echo -n "${json}" > "${QEMU_JSON_OUTPUT}"
