#!/usr/bin/env bash

set -euo pipefail
shopt -s inherit_errexit

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)
workspace_dir=$(realpath "${BUILD_WORKSPACE_DIRECTORY:-$(pwd)}")
lockfile_sha256="${workspace_dir}/image/mirror/SHA256SUMS"

DNF5=${DNF5:-dnf5}
DNF5=$(realpath "$(command -v "${DNF5}")")

AWS=${AWS:-aws}
AWS=$(realpath "$(command -v "${AWS}")")

DNF_CONF=${DNF_CONF:-${script_dir}/dnf.conf}
DNF_CONF=$(realpath "${DNF_CONF}")
PACKAGES=${PACKAGES:-${script_dir}/packages.txt}
PACKAGES=$(realpath "${PACKAGES}")
REPOSDIR=${REPOSDIR:-${script_dir}/upstream-repos}
if [[ ! -d ${REPOSDIR} ]]; then
  REPOSDIR=$(dirname "${REPOSDIR}")
fi
REPOSDIR=$(realpath "${REPOSDIR}")
OUTDIR="${OUTDIR:-$(mktemp -d)}"
OUTDIR=$(realpath "${OUTDIR}")

echo "Writing rpms to ${OUTDIR}"
echo "Lockfile location ${lockfile_sha256}"

lockfile_backup=$(mktemp)
lockfile_tmp=$(mktemp)
new_packages=$(mktemp)
cp "${lockfile_sha256}" "${lockfile_backup}"

download() {
  mkdir -p "${OUTDIR}"
  # shellcheck disable=SC2046
  "${DNF5}" \
    "--config=${DNF_CONF}" \
    "--setopt=reposdir=${REPOSDIR}" \
    "--releasever=40" \
    download \
    "--destdir=${OUTDIR}" \
    --resolve --alldeps \
    $(tr '\n' ' ' < "${PACKAGES}")
}

mirror() {
  local sha256
  local rpm
  while IFS="" read -r rpm; do
    rpm="${OUTDIR}/${rpm}"
    sha256=$(sha256sum "${rpm}" | cut -d' ' -f1)
    if "${AWS}" s3 ls "s3://cdn-constellation-backend/constellation/cas/sha256/${sha256}"; then
      echo "${sha256}" "${rpm}" already mirrored
    else
      echo "${sha256}" "${rpm}" mirroring...
      "${AWS}" s3 cp "${rpm}" "s3://cdn-constellation-backend/constellation/cas/sha256/${sha256}"
    fi
  done < "${new_packages}"
}

lockfile() {
  touch "${lockfile_tmp}"
  cd "${OUTDIR}" && sha256sum -- *.rpm > "${lockfile_tmp}" && cd -
}

overwrite_lockfile() {
  cp "${lockfile_tmp}" "${lockfile_sha256}"
}

diff_and_exit_if_lockfile_unchanged() {
  if cmp --silent "${lockfile_backup}" "${lockfile_tmp}"; then
    echo "No changes to lockfile"
    exit 0
  fi
  diff -Naur "${lockfile_backup}" "${lockfile_tmp}" || true
  comm -13 <(sort "${lockfile_backup}") <(sort "${lockfile_tmp}") | cut -d' ' -f3 > "${new_packages}"
  echo "New packages:"
  cat "${new_packages}"
}

download
lockfile
diff_and_exit_if_lockfile_unchanged
mirror
overwrite_lockfile
