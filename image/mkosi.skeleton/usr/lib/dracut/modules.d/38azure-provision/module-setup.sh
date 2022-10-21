#!/usr/bin/env bash
# Copyright (c) Edgeless Systems GmbH
#
# SPDX-License-Identifier: AGPL-3.0-only

depends() {
    echo systemd
}

install_and_enable_unit() {
    unit="$1"; shift
    target="$1"; shift
    inst_simple "$moddir/$unit" "$systemdsystemunitdir/$unit"
    mkdir -p "${initdir}${systemdsystemconfdir}/${target}.wants"
    ln_r "${systemdsystemunitdir}/${unit}" \
        "${systemdsystemconfdir}/${target}.wants/${unit}"
}

install() {
    inst_multiple \
        bash \
        curl \
        grep \
        sed

    inst_script "$moddir/azure-provisioning.sh" \
        "/usr/local/bin/azure-provisioning"
    install_and_enable_unit "azure-provisioning.service" \
        "basic.target"
}
