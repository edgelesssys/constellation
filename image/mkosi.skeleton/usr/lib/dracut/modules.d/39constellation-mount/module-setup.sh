#!/usr/bin/env bash
# Copyright (c) Edgeless Systems GmbH
#
# SPDX-License-Identifier: AGPL-3.0-only

# Note: This script is sourced.

depends() {
  # systemd-network-management expands to: systemd systemd-hostnamed systemd-networkd systemd-resolved systemd-timedated systemd-timesyncd
  echo dracut-systemd systemd-network-management systemd-veritysetup systemd-udevd
  return 0
}

install_and_enable_unit() {
  unit="$1"
  shift
  target="$1"
  shift
  inst_simple "${moddir:?}/${unit}" "${systemdsystemunitdir:?}/${unit}"
  mkdir -p "${initdir:?}${systemdsystemconfdir:?}/${target}.wants"
  ln_r "${systemdsystemunitdir}/${unit}" \
    "${systemdsystemconfdir}/${target}.wants/${unit}"
}

install_path() {
  local dir="$1"
  shift
  mkdir -p "${initdir}/${dir}"
}

install() {
  inst_multiple \
    bash
  inst_script "/usr/sbin/disk-mapper" \
    "/usr/sbin/disk-mapper"

  inst_script "${moddir}/prepare-state-disk.sh" \
    "/usr/sbin/prepare-state-disk"
  install_and_enable_unit "prepare-state-disk.service" \
    "basic.target"
  install_and_enable_unit "configure-constel-csp.service" \
    "basic.target"

  # aws nvme disks
  inst_multiple \
    tail \
    tr \
    head

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

  inst_script "${moddir}/aws-nvme-disk.sh" \
    "/usr/sbin/aws-nvme-disk"
  install_and_enable_unit "aws-nvme-disk.service" \
    "basic.target"

  # TLS / CA store in initramfs
  install_path /etc/pki/tls/certs/
  inst_simple /etc/pki/tls/certs/ca-bundle.crt \
    /etc/pki/tls/certs/ca-bundle.crt

  # backport of https://github.com/dracutdevs/dracut/commit/dcbe23c14d13ca335ad327b7bb985071ca442f12
  inst_simple "${moddir}/sysusers-dracut.conf" "${systemdsystemunitdir}/systemd-sysusers.service.d/sysusers-dracut.conf"
  inst_simple /usr/lib/systemd/resolved.conf.d/fallback_dns.conf \
    /usr/lib/systemd/resolved.conf.d/fallback_dns.conf
}
