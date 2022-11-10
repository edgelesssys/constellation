#!/usr/bin/env bash
# Copyright (c) Edgeless Systems GmbH
#
# SPDX-License-Identifier: AGPL-3.0-only

# This script contains shared functions for pcr calculation.

set -euo pipefail
shopt -s inherit_errexit

pcr_extend() {
    local CURRENT_PCR="$1"
    local EXTEND_WITH="$2"
    local HASH_FUNCTION="$3"
    ( echo -n "${CURRENT_PCR}" | xxd -r -p ; echo -n "${EXTEND_WITH}" | xxd -r -p; ) | ${HASH_FUNCTION} | cut -d " " -f 1
}

extract () {
    local image="$1"
    local path="$2"
    local output="$3"
    sudo systemd-dissect --copy-from "${image}" "${path}" "${output}"
}

mktempdir () {
    mktemp -d
}

cleanup () {
    local dir="$1"
    rm -rf "${dir}"
}
