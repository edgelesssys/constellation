#!/usr/bin/env bash
# Copyright (c) Edgeless Systems GmbH
#
# SPDX-License-Identifier: AGPL-3.0-only

set -euo pipefail
shopt -s extglob nullglob

AWS_STATE_DISK_SYMLINK="/dev/sdb"

# hack: aws nvme udev rules are never executed. Create symlinks for the nvme devices manually.
while [ ! -L "${AWS_STATE_DISK_SYMLINK}" ]
do
    for nvmedisk in /dev/nvme+([0-9])
    do
        linkname=$(nvme admin-passthru --opcode=0x06 --cdw10=1 -b --data-len=4096 -r "${nvmedisk}" | tail -c +3072 | tr -d ' ') || true
        if [ -n "${linkname}" ]; then
            ln -s "${nvmedisk}" "/dev/${linkname}"
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
sleep 2
