#!/usr/bin/env bash
# Copyright (c) Edgeless Systems GmbH
#
# SPDX-License-Identifier: AGPL-3.0-only

# This script is used to add a signed shim to the image.raw file EFI partition after running `mkosi build`.

set -euo pipefail
shopt -s inherit_errexit

if (($# != 1)); then
  echo "Usage: $0 <image.raw>"
  exit 1
fi

# SOURCE is the URL used to download the signed shim RPM
SOURCE=https://kojipkgs.fedoraproject.org/packages/shim/15.6/2/x86_64/shim-x64-15.6-2.x86_64.rpm
# EXPECTED_SHA512 is the SHA512 checksum of the signed shim RPM
EXPECTED_SHA512=971978bddee95a6a134ef05c4d88cf5df41926e631de863b74ef772307f3e106c82c8f6889c18280d47187986abd774d8671c5be4b85b1b0bb3d1858b65d02cf
TMPDIR=$(mktemp -d)

pushd "${TMPDIR}"

curl -sL -o shim.rpm "${SOURCE}"
echo "Checking SHA512 checksum of signed shim..."
sha512sum -c <<< "${EXPECTED_SHA512}  shim.rpm"
rpm2cpio shim.rpm | cpio -idmv
echo "${TMPDIR}"

popd

MOUNTPOINT=$(mktemp -d)
sectoroffset=$(sfdisk -J "${1}" | jq -r '.partitiontable.partitions[0].start')
byteoffset=$((sectoroffset * 512))
mount -o offset="${byteoffset}" "${1}" "${MOUNTPOINT}"

mkdir -p "${MOUNTPOINT}/EFI/BOOT/"
cp "${TMPDIR}/boot/efi/EFI/BOOT/BOOTX64.EFI" "${MOUNTPOINT}/EFI/BOOT/"
cp "${TMPDIR}/boot/efi/EFI/fedora/mmx64.efi" "${MOUNTPOINT}/EFI/BOOT/"
cp "${MOUNTPOINT}/EFI/systemd/systemd-bootx64.efi" "${MOUNTPOINT}/EFI/BOOT/grubx64.efi"

# Remove unused kernel and initramfs from EFI to save space
# We boot from unified kernel image anyway
rm -f "${MOUNTPOINT}"/*/*/{linux,initrd}

umount "${MOUNTPOINT}"
rm -rf "${MOUNTPOINT}"
rm -rf "${TMPDIR}"
