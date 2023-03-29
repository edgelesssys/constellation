#!/usr/bin/env bash

set -euo pipefail
set -o errtrace
shopt -s inherit_errexit

echo "Pulling cert-manager Helm chart..."

function cleanup {
  rm -r "charts/cert-manager/README.md" "charts/cert-manager-v1.10.0.tgz"
}

trap cleanup EXIT

helm pull cert-manager \
  --version 1.10.0 \
  --repo "https://charts.jetstack.io" \
  --untar \
  --untardir "charts"

echo # final newline
