#!/usr/bin/env bash

set -euo pipefail
set -o errtrace
shopt -s inherit_errexit

echo "Pulling cert-manager Helm chart..."
version="1.12.6"

function cleanup {
  rm -r "charts/cert-manager/README.md" "charts/cert-manager-v${version}.tgz"
}

trap cleanup EXIT

helm pull cert-manager \
  --version ${version} \
  --repo "https://charts.jetstack.io" \
  --untar \
  --untardir "charts"

get_sha256_hash() {
  local component="$1"
  local url="https://quay.io/v2/jetstack/${component}/manifests/v${version}"
  local hash
  has_error=$(curl -s -L -H "Accept: application/vnd.docker.distribution.manifest.v2+json" "$url" | jq 'has("errors")')

  # Check if 'errors' attribute exists
  if [ "$has_error" = "true" ]; then
    echo "Errors attribute found. Exiting with status 1."
    exit 1
  fi

  curl -s -L -H "Accept: application/vnd.docker.distribution.manifest.v2+json" "$url" | sha256sum | awk '{print $1}'
}

echo "Pinning cert-manager images..."
yq=$(realpath @@YQ@@)
stat "${yq}" >> /dev/null
v=$(get_sha256_hash "cert-manager-controller")
${yq} eval -i '.image.digest = "sha256:'"$v"'"' charts/cert-manager/values.yaml

v=$(get_sha256_hash "cert-manager-webhook")
${yq} eval -i '.webhook.image.digest = "sha256:'"$v"'"' charts/cert-manager/values.yaml

v=$(get_sha256_hash "cert-manager-cainjector")
${yq} eval -i '.cainjector.image.digest = "sha256:'"$v"'"' charts/cert-manager/values.yaml

v=$(get_sha256_hash "cert-manager-acmesolver")
${yq} eval -i '.acmesolver.image.digest = "sha256:'"$v"'"' charts/cert-manager/values.yaml

v=$(get_sha256_hash "cert-manager-ctl")
${yq} eval -i '.startupapicheck.image.digest = "sha256:'"$v"'"' charts/cert-manager/values.yaml

echo # final newline
