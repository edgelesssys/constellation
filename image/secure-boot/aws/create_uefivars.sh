#!/usr/bin/env bash
set -euo pipefail

TMPDIR=$(mktemp -d /tmp/uefivars-XXXXXXXXXXXXXX)
git clone https://github.com/awslabs/python-uefivars ${TMPDIR}

"${TMPDIR}/uefivars.py" -i none -o aws -O "$1" -P ${PKI}/PK.esl -K ${PKI}/KEK.esl --db ${PKI}/db.esl

rm -rf "${TMPDIR}"
