#!/usr/bin/env bash
# Copyright (c) Edgeless Systems GmbH
#
# SPDX-License-Identifier: AGPL-3.0-only

set -euo pipefail
shopt -s inherit_errexit

# Show progress on pipes if `pv` is installed
# Otherwise use plain cat
if ! command -v pv &> /dev/null; then
  PV="cat"
else
  PV="pv"
fi

pack() {
  local cloudprovider=$1
  local unpacked_image=$2
  local packed_image=$3
  local unpacked_image_dir
  unpacked_image_dir=$(mktemp -d)
  local unpacked_image_filename
  unpacked_image_filename=disk.raw
  local tmp_tar_file
  tmp_tar_file=$(mktemp -t verity.XXXXXX.tar)
  cp "${unpacked_image}" "${unpacked_image_dir}/${unpacked_image_filename}"

  case ${cloudprovider} in

  gcp)
    echo "📥 Packing GCP image..."
    tar --owner=0 --group=0 -C "${unpacked_image_dir}" -Sch --format=oldgnu -f "${tmp_tar_file}" "${unpacked_image_filename}"
    "${PV}" "${tmp_tar_file}" | pigz -9c > "${packed_image}"
    rm "${tmp_tar_file}"
    echo "  Repacked image stored in ${packed_image}"
    ;;

  azure)
    echo "📥 Packing Azure image..."
    # Disk Images on Azure have to be a multiple of 1MiB in size.
    # Marketplace images additionally have to be at least 1GiB in size.
    truncate -s %2G "${unpacked_image_dir}/${unpacked_image_filename}" # at least 1GiB
    # truncate -s +1M "${unpacked_image_dir}/${unpacked_image_filename}" # ensure > 1 GiB, but multiple of 1MiB
    qemu-img convert -p -f raw -O vpc -o force_size,subformat=fixed "${unpacked_image_dir}/${unpacked_image_filename}" "${packed_image}"
    echo "  Repacked image stored in ${packed_image}"
    ;;

  *)
    echo "unknown cloud provider"
    exit 1
    ;;
  esac

  rm -r "${unpacked_image_dir}"

}

if [[ $# -ne 3 ]]; then
  echo "Usage: $0 <cloudprovider> <unpacked_image> <packed_image>"
  exit 1
fi

pack "${1}" "${2}" "${3}"
