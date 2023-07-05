#!/usr/bin/env bash

set -euo pipefail
shopt -s inherit_errexit

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)
REPOSITORY_ROOT="${REPOSITORY_ROOT:-$(cd "${SCRIPT_DIR}" && git rev-parse --show-toplevel)}"

goos() {
  case "$(uname -sr)" in
  Darwin*) echo 'darwin' ;;
  Linux*) echo 'linux' ;;
  *)
    echo 'Unknown OS' >&2
    exit 1
    ;;
  esac
}

goarch() {
  case $(uname -m) in
  x86_64) echo 'amd64' ;;
  arm) echo 'arm64' ;; # this is slightly simplified, but we only care about arm64
  arm64) echo 'arm64' ;;
  *)
    echo 'Unknown arch' >&2
    exit 1
    ;;
  esac
}

timestamp() {
  git show -s --date=format:'%Y-%m-%dT%H:%M:%S' --format=%cd HEAD
}

stamp_version() {
  local version
  version=$(fixed_version)
  # shellcheck disable=SC2310
  if is_pre_version; then
    version=$(pseudo_version)
  fi
  remove_v_prefix "${version}"
}

is_pre_version() {
  local version
  version=$(cat "${REPOSITORY_ROOT}/version.txt")
  [[ ${version} =~ ^.*-pre.*$ ]]
}

remove_v_prefix() {
  local version=$1
  echo "${version#v}"
}

# pseudo_version is a bash implementation of the go pseudo version format
# We only care about pre-release versions, so we can simplify the implementation
# See https://pkg.go.dev/golang.org/x/mod/module#PseudoVersion
pseudo_version() {
  local prefix
  prefix=$(fixed_version)
  echo "${prefix}.0.$(git show -s --date=format:'%Y%m%d%H%M%S' --format=%cd HEAD)-$(git rev-parse --short=12 HEAD)"
}

fixed_version() {
  cat "${REPOSITORY_ROOT}/version.txt"
}

echo "REPO_URL https://github.com/edgelesssys/constellation.git"
echo "STABLE_STAMP_COMMIT $(git rev-parse HEAD)"
echo "STABLE_STAMP_STATE $(git update-index -q --really-refresh && git diff-index --quiet HEAD -- && echo "clean" || echo "dirty")"
echo "STABLE_STAMP_VERSION $(stamp_version)"
echo "STABLE_STAMP_TIME $(timestamp)"
