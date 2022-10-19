#!/bin/bash
set -euo pipefail
shopt -s extglob nullglob

GCP_STATE_DISK_SYMLINK="/dev/disk/by-id/google-state-disk"

# hack: google nvme udev rules are never executed. Create symlinks for the nvme devices manually.
while [ ! -L "${GCP_STATE_DISK_SYMLINK}" ]
do
    for nvmedisk in /dev/nvme0n+([0-9])
    do
        /usr/lib/udev/google_nvme_id -s -d "${nvmedisk}" || true
    done
    if [ -L "${GCP_STATE_DISK_SYMLINK}" ]; then
        break
    fi
    echo "Waiting for state disk to appear.."
    sleep 2
done

echo "Google state disk found"
echo ${GCP_STATE_DISK_SYMLINK} → $(readlink -f "${GCP_STATE_DISK_SYMLINK}")
sleep 2
