#!/usr/bin/env bash

set -euo pipefail
shopt -s inherit_errexit

platforms=(
  darwin_amd64
  darwin_arm64
  linux_amd64
  linux_arm64
)
bucket=cdn-constellation-backend

dir=$(mktemp -d -t constellation-XXXXXXXXXX)
trap 'rm -rf "${dir}"' EXIT

bazel build --config nostamp "//hack/pseudo-version:all"
workspace_dir=$(git rev-parse --show-toplevel)

for platform in "${platforms[@]}"; do
  echo "Building for ${platform}..."
  target="//hack/pseudo-version:pseudo_version_${platform}"
  cp "$(bazel cquery --config nostamp --output=files "${target}")" "${dir}/pseudo_version_${platform}"
  sha256="$(sha256sum "${dir}/pseudo_version_${platform}" | cut -d ' ' -f 1)"
  echo "${platform} ${sha256}" | tee -a "${dir}/checksums.txt"
  aws s3 cp "${dir}/pseudo_version_${platform}" "s3://${bucket}/constellation/cas/sha256/${sha256}"
  echo "${sha256}" > "${workspace_dir}/tools/pseudo_version_${platform}.sha256"
done

cat "${dir}/checksums.txt"
