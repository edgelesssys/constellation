#!/usr/bin/env bash
# Copyright (c) Edgeless Systems GmbH
#
# SPDX-License-Identifier: AGPL-3.0-only

set -euo pipefail
shopt -s extglob nullglob

AWS_STATE_DISK_DEVICENAME="sdb"
AWS_STATE_DISK_SYMLINK="/dev/${AWS_STATE_DISK_DEVICENAME}"

# hack: aws nvme udev rules are never executed. Create symlinks for the nvme devices manually.
while [ ! -L "${AWS_STATE_DISK_SYMLINK}" ]
do
    for nvmedisk in /dev/nvme*n1
    do
        linkname=$(nvme amzn id-ctrl -b "${nvmedisk}" | tail -c +3072 | tr -d ' ') || true
        if [ -n "${linkname}" ] && [ "${linkname}" == "${AWS_STATE_DISK_DEVICENAME}" ]; then
            ln -s "${nvmedisk}" "${AWS_STATE_DISK_SYMLINK}"
        fi
    done
    if [ -L "${AWS_STATE_DISK_SYMLINK}" ]; then
        break
    fi
    echo "Waiting for state disk to appear.."
    sleep 2
done

echo "AWS state disk found"
echo ${AWS_STATE_DISK_SYMLINK} → $(readlink -f "${AWS_STATE_DISK_SYMLINK}")
