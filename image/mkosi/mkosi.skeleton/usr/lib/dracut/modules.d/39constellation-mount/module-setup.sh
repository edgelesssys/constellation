#!/bin/bash

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
        bash
    inst_script "/usr/sbin/disk-mapper" \
        "/usr/sbin/disk-mapper"

    inst_script "$moddir/prepare-state-disk.sh" \
        "/usr/sbin/prepare-state-disk"
    install_and_enable_unit "prepare-state-disk.service" \
        "basic.target"
    inst_script "$moddir/google-nvme-disk.sh" \
        "/usr/sbin/google-nvme-disk"
    install_and_enable_unit "google-nvme-disk.service" \
        "basic.target"
    install_and_enable_unit "configure-constel-csp.service" \
        "basic.target"

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
    inst_simple "/usr/lib/udev/rules.d/64-gce-disk-removal.rules" \
        "/usr/lib/udev/rules.d/64-gce-disk-removal.rules"
    inst_simple "/usr/lib/udev/rules.d/65-gce-disk-naming.rules" \
        "/usr/lib/udev/rules.d/65-gce-disk-naming.rules"
}
