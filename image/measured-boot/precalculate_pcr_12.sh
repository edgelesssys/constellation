#!/usr/bin/env bash
# Copyright (c) Edgeless Systems GmbH
#
# SPDX-License-Identifier: AGPL-3.0-only

# This script is used to precalculate the PCR[12] value for a Constellation OS image.
# PCR[12] contains the hash of the kernel command line and is measured by systemd-boot.
# This value was previously measured into PCR[8].
# This script may produce wrong results for systemd-boot versions < 251.
# Usage: precalculate_pcr_12.sh <path to image> <path to output file> <csp>

set -euo pipefail
shopt -s inherit_errexit
source "$(dirname "$0")/measure_util.sh"

get_cmdline_from_uki() {
  local uki="$1"
  local output="$2"
  objcopy -O binary --only-section=.cmdline "${uki}" "${output}"
}

cmdline_measure() {
  local path="$1"
  local tmp
  tmp=$(mktemp)
  # convert to utf-16le
  iconv -f utf-8 -t utf-16le "${path}" -o "${tmp}"
  sha256sum "${tmp}" | cut -d " " -f 1
  rm "${tmp}"
}

write_output() {
  local out="$1"
  cat > "${out}" << EOF
{
  "measurements": {
    "12": {
      "expected": "${expected_pcr_12}"
    }
  },
  "cmdline": "${cmdline}",
  "cmdline-sha256": "${cmdline_hash}"
}
EOF
}

IMAGE="$1"
OUT="$2"
CSP="$3"

DIR=$(mktempdir)
trap 'cleanup "${DIR}"' EXIT

extract "${IMAGE}" "/efi/EFI/Linux" "${DIR}/uki"
sudo chown -R "${USER}:${USER}" "${DIR}/uki"
cp "${DIR}"/uki/*.efi "${DIR}/03-uki.efi"
get_cmdline_from_uki "${DIR}/03-uki.efi" "${DIR}/cmdline"
cmdline=$(cat "${DIR}/cmdline")

cmdline_hash=$(cmdline_measure "${DIR}/cmdline")
cleanup "${DIR}"

expected_pcr_12=0000000000000000000000000000000000000000000000000000000000000000
expected_pcr_12=$(pcr_extend "${expected_pcr_12}" "${cmdline_hash}" "sha256sum")
if [[ ${CSP} == "azure" ]]; then
  # Azure displays the boot menu
  # triggering an extra measurement of the kernel command line.
  expected_pcr_12=$(pcr_extend "${expected_pcr_12}" "${cmdline_hash}" "sha256sum")
fi

echo "Kernel commandline:            ${cmdline}"
echo "Kernel Commandline measurement ${cmdline_hash}"
echo ""
echo "Expected PCR[12]:               ${expected_pcr_12}"
echo ""

write_output "${OUT}"
