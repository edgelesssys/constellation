#!/usr/bin/env bash
# Copyright (c) Edgeless Systems GmbH
#
# SPDX-License-Identifier: AGPL-3.0-only

depends() {
    # expands to: systemd systemd-hostnamed systemd-networkd systemd-resolved systemd-timedated systemd-timesyncd
    echo systemd-network-management
}

install_and_enable_unit() {
    unit="$1"; shift
    target="$1"; shift
    inst_simple "$moddir/$unit" "$systemdsystemunitdir/$unit"
    mkdir -p "${initdir}${systemdsystemconfdir}/${target}.wants"
    ln_r "${systemdsystemunitdir}/${unit}" \
        "${systemdsystemconfdir}/${target}.wants/${unit}"
}

install_dir_recursive() {
    dir="$1"; shift
    insert_location=$(dirname "${initdir}${dir}")
    mkdir -p "${insert_location}"
    cp -r "${dir}" "${insert_location}/"
}

install() {
    inst_multiple \
        bash
    inst_script "/usr/sbin/disk-mapper" \
        "/usr/sbin/disk-mapper"

    inst_script "$moddir/prepare-state-disk.sh" \
        "/usr/sbin/prepare-state-disk"
    install_and_enable_unit "prepare-state-disk.service" \
        "basic.target"
    install_and_enable_unit "configure-constel-csp.service" \
        "basic.target"

    # aws nvme disks
    inst_multiple \
        tail \
        tr

    # azure scsi disks
    inst_multiple \
        cut \
        readlink

    # gcp nvme disks
    inst_multiple \
        date \
        xxd \
        grep \
        sed \
        ln \
        command \
        readlink

    inst_script "/usr/sbin/nvme" \
        "/usr/sbin/nvme"
    inst_script "/usr/lib/udev/google_nvme_id" \
        "/usr/lib/udev/google_nvme_id"
    inst_rules "64-gce-disk-removal.rules" "65-gce-disk-naming.rules"

    inst_script "$moddir/aws-nvme-disk.sh" \
        "/usr/sbin/aws-nvme-disk"
    install_and_enable_unit "aws-nvme-disk.service" \
        "basic.target"

    # TLS / CA store in initramfs
    install_dir_recursive /etc/pki
    install_dir_recursive /usr/share/pki
}
