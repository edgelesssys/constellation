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

path="constellation/v1/ref/${REF}/stream/${STREAM}/${IMAGE_VERSION}/image/csp/openstack/image.raw"
aws s3 cp "${OPENSTACK_IMAGE_PATH}" "s3://${OPENSTACK_BUCKET}/${path}" --no-progress

image_url="${OPENSTACK_BASE_URL}/${path}"

json=$(jq -ncS \
  --arg image_url "${image_url}" \
  '{"openstack": {"sev": $image_url}}')
echo -n "${json}" > "${OPENSTACK_JSON_OUTPUT}"
