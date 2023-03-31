#!/usr/bin/env bash

set -euo pipefail
shopt -s inherit_errexit

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)
REPOSITORY_ROOT="${REPOSITORY_ROOT:-$(cd "${SCRIPT_DIR}" && git rev-parse --show-toplevel)}"

ensure_pseudo_version_tool() {
  if [[ ! -f "${REPOSITORY_ROOT}/tools/pseudo-version" ]]; then
    go build -o "${REPOSITORY_ROOT}/tools/pseudo-version" "${REPOSITORY_ROOT}"/hack/pseudo-version >&2
  fi
}

pseudo_version() {
  ensure_pseudo_version_tool
  "${REPOSITORY_ROOT}/tools/pseudo-version" -skip-v
}

timestamp() {
  ensure_pseudo_version_tool
  "${REPOSITORY_ROOT}/tools/pseudo-version" -print-timestamp -timestamp-format '2006-01-02T15:04:05Z07:00'
}

echo "REPO_URL https://github.com/edgelesssys/constellation.git"
echo "STABLE_STAMP_COMMIT $(git rev-parse HEAD)"
echo "STABLE_STAMP_STATE $(git update-index -q --really-refresh && git diff-index --quiet HEAD -- && echo "clean" || echo "dirty")"
echo "STABLE_STAMP_VERSION $(pseudo_version)"
echo "STABLE_STAMP_TIME $(timestamp)"
