#!/usr/bin/env bash

set -euo pipefail
set -o errtrace
shopt -s inherit_errexit

echo "Pulling cert-manager Helm chart..."
version="1.15.0"

function cleanup {
  rm -rf "charts/cert-manager/README.md" "charts/cert-manager-v${version}.tgz"
}

trap cleanup EXIT

helm pull cert-manager \
  --version "${version}" \
  --repo "https://charts.jetstack.io" \
  --untar \
  --untardir "charts"

get_sha256_hash() {
  local component="$1"
  local url="https://quay.io/v2/jetstack/${component}/manifests/v${version}"
  curl -fsSL -H "Accept: application/vnd.docker.distribution.manifest.v2+json" "${url}" | sha256sum | awk '{print $1}'
}

echo "Pinning cert-manager images..."
v=$(get_sha256_hash "cert-manager-controller")
yq eval -i '.image.digest = "sha256:'"${v}"'"' charts/cert-manager/values.yaml

v=$(get_sha256_hash "cert-manager-webhook")
yq eval -i '.webhook.image.digest = "sha256:'"${v}"'"' charts/cert-manager/values.yaml

v=$(get_sha256_hash "cert-manager-cainjector")
yq eval -i '.cainjector.image.digest = "sha256:'"${v}"'"' charts/cert-manager/values.yaml

v=$(get_sha256_hash "cert-manager-acmesolver")
yq eval -i '.acmesolver.image.digest = "sha256:'"${v}"'"' charts/cert-manager/values.yaml

v=$(get_sha256_hash "cert-manager-startupapicheck")
yq eval -i '.startupapicheck.image.digest = "sha256:'"${v}"'"' charts/cert-manager/values.yaml

echo # final newline
