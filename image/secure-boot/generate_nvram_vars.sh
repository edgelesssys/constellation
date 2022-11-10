#!/usr/bin/env bash
# Copyright (c) Edgeless Systems GmbH
#
# SPDX-License-Identifier: AGPL-3.0-only

set -euo pipefail
shopt -s inherit_errexit

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
BASE_DIR=$(realpath "${SCRIPT_DIR}/..")

# Set to qemu+tcp://localhost:16599/system for dockerized libvirt setup
if [[ -z "${LIBVIRT_SOCK}" ]]; then
    LIBVIRT_SOCK=qemu:///system
fi

libvirt_nvram_gen () {
    local image_path="${1}"
    if test -f "${BASE_DIR}/image.nvram.template"; then
        echo "NVRAM template already generated: $(realpath "--relative-to=$(pwd)" "${BASE_DIR}"/image.nvram.template)"
        return
    fi
    if ! test -f "${image_path}"; then
        echo "Image \"${image_path}\" does not exist yet. To generate nvram, create disk image first."
        return
    fi

    OVMF_CODE=/usr/share/OVMF/OVMF_CODE_4M.ms.fd
    OVMF_VARS=/usr/share/OVMF/OVMF_VARS_4M.ms.fd
    if ! test -f "${OVMF_CODE}"; then
        OVMF_CODE=/usr/share/OVMF/OVMF_CODE.secboot.fd
    fi
    if ! test -f "${OVMF_VARS}"; then
        OVMF_VARS=/usr/share/OVMF/OVMF_VARS.secboot.fd
    fi

    echo "Using OVMF_CODE: ${OVMF_CODE}"
    echo "Using OVMF_VARS: ${OVMF_VARS}"

    # generate nvram file using libvirt
    virt-install --name constell-nvram-gen \
        --connect "${LIBVIRT_SOCK}" \
        --nonetworks \
        --description 'Constellation' \
        --ram 1024 \
        --vcpus 1 \
        --osinfo detect=on,require=off \
        --disk "${image_path},format=raw" \
        --boot "machine=q35,menu=on,loader=${OVMF_CODE},loader.readonly=yes,loader.type=pflash,nvram.template=${OVMF_VARS},nvram=${BASE_DIR}/image.nvram,loader_secure=yes" \
        --features smm.state=on \
        --noautoconsole
    echo -e 'connect using'
    echo -e '    \u001b[1mvirsh console constell-nvram-gen\u001b[0m'
    echo -e ''
    echo -e 'Load db cert with MokManager or enroll full PKI with firmware setup'
    echo -e ''
    echo -e '    \u001b[1mMokManager\u001b[0m'
    echo -e '    For mokmanager, try to boot as usual. You will see this message:'
    echo -e '    > "Verification failed: (0x1A) Security Violation"'
    echo -e '    Press OK, then ENTER, then "Enroll key from disk"'
    echo -e '    Select the following key:'
    echo -e '    > \u001b[1m/EFI/loader/keys/auto/db.cer\u001b[0m'
    echo -e '    Press Continue, then choose "Yes" to the question "Enroll the key(s)?"'
    echo -e '    Choose reboot and continue this script.'
    echo -e ''
    echo -e '    \u001b[1mFirmware setup\u001b[0m'
    echo -e '    For firmware setup, press F2.'
    echo -e '    Go to "Device Manager">"Secure Boot Configuration">"Secure Boot Mode"'
    echo -e '    Choose "Custom Mode"'
    echo -e '    Go to "Custom Securee Boot Options"'
    echo -e '    Go to "PK Options">"Enroll PK", Press "Y" if queried, "Enroll PK using File"'
    echo -e '    Select the following cert: \u001b[1m/EFI/loader/keys/auto/PK.cer\u001b[0m'
    echo -e '    Choose "Commit Changes and Exit"'
    echo -e '    Go to "KEK Options">"Enroll KEK", Press "Y" if queried, "Enroll KEK using File"'
    echo -e '    Select the following cert: \u001b[1m/EFI/loader/keys/auto/KEK.cer\u001b[0m'
    echo -e '    Choose "Commit Changes and Exit"'
    echo -e '    Go to "DB Options">"Enroll Signature">"Enroll Signature using File"'
    echo -e '    Select the following cert: \u001b[1m/EFI/loader/keys/auto/db.cer\u001b[0m'
    echo -e '    Choose "Commit Changes and Exit"'
    echo -e '    Repeat the last step for the following certs:'
    echo -e '    > \u001b[1m/EFI/loader/keys/auto/MicWinProPCA2011_2011-10-19.crt\u001b[0m'
    echo -e '    > \u001b[1m/EFI/loader/keys/auto/MicCorUEFCA2011_2011-06-27.crt\u001b[0m'
    echo -e '    Reboot and continue this script.'
    echo -e ''
    echo -e 'Press ENTER to continue after you followed one of the guides from above.'
    read -r
    sudo cp "${BASE_DIR}/image.nvram" "${BASE_DIR}/image.nvram.template"
    virsh --connect "${LIBVIRT_SOCK}" destroy --domain constell-nvram-gen
    virsh --connect "${LIBVIRT_SOCK}" undefine --nvram constell-nvram-gen
    rm -f "${BASE_DIR}/image.nvram"

    echo "NVRAM template generated: $(realpath "--relative-to=$(pwd)" "${BASE_DIR}"/image.nvram.template)"
}

libvirt_nvram_gen "$1"
