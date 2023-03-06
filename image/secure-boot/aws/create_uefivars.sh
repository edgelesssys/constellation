#!/usr/bin/env bash

set -euo pipefail
shopt -s inherit_errexit

TMPDIR=$(mktemp -d /tmp/uefivars-XXXXXXXXXXXXXX)
git clone --branch v1.0.0 https://github.com/awslabs/python-uefivars "${TMPDIR}"
cd "${TMPDIR}" && git reset 9679002a4392d8e7831d2dbda3fab41ccc5c6b8c --hard

"${TMPDIR}/uefivars.py" -i none -o aws -O "$1" -P "${PKI}"/PK.esl -K "${PKI}"/KEK.esl --db "${PKI}"/db.esl

rm -rf "${TMPDIR}"
