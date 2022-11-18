#!/usr/bin/env bash
# Copyright (c) Edgeless Systems GmbH
#
# SPDX-License-Identifier: AGPL-3.0-only

# Usage: find-image.sh

set -euo pipefail
shopt -s inherit_errexit

base_url="https://cdn.confidential.cloud"
bucket="cdn-constellation-backend"
newest_debug_image_path=$(aws s3api list-objects-v2 \
  --output text \
  --bucket "${bucket}" \
  --prefix constellation/v1/images/debug-v \
  --query "reverse(sort_by(Contents, &LastModified))[0].Key")

image_version_uid=$(basename "${newest_debug_image_path}" .json)
url="${base_url}/${newest_debug_image_path}"
echo "Found image version UID:"
echo "${image_version_uid}"

echo "Containing the following images:"
echo "${url}"
curl -sL "${url}" | jq
