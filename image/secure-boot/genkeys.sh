#!/usr/bin/env bash
# Copyright (c) Edgeless Systems GmbH
#
# SPDX-License-Identifier: AGPL-3.0-only

# This script generates a PKI for secure boot.
# It is based on the example from https://github.com/systemd/systemd/blob/main/man/loader.conf.xml
# This is meant to be used for development purposes only.
# Release images are signed using a different set of keys.
# Set PKI to an empty folder and PKI_SET to "dev".

set -euo pipefail
shopt -s inherit_errexit

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
TEMPLATES=${SCRIPT_DIR}/templates
BASE_DIR=$(realpath "${SCRIPT_DIR}/..")
if [[ -z "${PKI}" ]]; then
    PKI=${BASE_DIR}/pki
fi
if [[ -z "${PKI_SET}" ]]; then
    PKI_SET=dev
fi

gen_pki () {
    # Only use for non-production images.
    # Use real PKI for production images instead.
    count=$(find "${PKI}" -maxdepth 1 \( -name '*.key' -o -name '*.crt' -o -name '*.cer' -o -name '*.esl' -o -name '*.auth' \) 2>/dev/null | wc -l)
    if [[ "${count}" != 0 ]]
    then
        echo PKI files "$(ls -1 "$(realpath "--relative-to=$(pwd)" "${PKI}")"/*.{key,crt,cer,esl,auth})" already exist
        return
    fi
    mkdir -p "${PKI}"
    pushd "${PKI}" || exit 1

    uuid=$(systemd-id128 new --uuid)
    for key in PK KEK db; do
        openssl req -new -x509 -config "${TEMPLATES}/${PKI_SET}_${key}.conf" -keyout "${key}.key" -out "${key}.crt" -nodes
        openssl x509 -outform DER -in "${key}.crt" -out "${key}.cer"
        cert-to-efi-sig-list -g "${uuid}" "${key}.crt" "${key}.esl"
    done

    for key in MicWinProPCA2011_2011-10-19.crt MicCorUEFCA2011_2011-06-27.crt MicCorKEKCA2011_2011-06-24.crt; do
        curl -sL "https://www.microsoft.com/pkiops/certs/${key}" --output "${key}"
        sbsiglist --owner 77fa9abd-0359-4d32-bd60-28f4e78f784b --type x509 --output "${key%crt}esl" "${key}"
    done

    # Optionally add Microsoft Windows Production CA 2011 (needed to boot into Windows).
    cat MicWinProPCA2011_2011-10-19.esl >> db.esl

    # Optionally add Microsoft Corporation UEFI CA 2011 (for firmware drivers / option ROMs
    # and third-party boot loaders (including shim). This is highly recommended on real
    # hardware as not including this may soft-brick your device (see next paragraph).
    cat MicCorUEFCA2011_2011-06-27.esl >> db.esl

    # Optionally add Microsoft Corporation KEK CA 2011. Recommended if either of the
    # Microsoft keys is used as the official UEFI revocation database is signed with this
    # key. The revocation database can be updated with [fwupdmgr(1)](https://www.freedesktop.org/software/systemd/man/fwupdmgr.html#).
    cat MicCorKEKCA2011_2011-06-24.esl >> KEK.esl

    sign-efi-sig-list -c PK.crt -k PK.key PK PK.esl PK.auth
    sign-efi-sig-list -c PK.crt -k PK.key KEK KEK.esl KEK.auth
    sign-efi-sig-list -c KEK.crt -k KEK.key db db.esl db.auth

    popd || exit 1
}

# gen_pki generates a PKI for testing purposes only.
# if keys/certs are already present in the pki folder, they are not regenerated.
gen_pki
