#!/usr/bin/env bash
# Copyright (c) Edgeless Systems GmbH
#
# SPDX-License-Identifier: AGPL-3.0-only

set -euo pipefail

if [ -z "${CONFIG_FILE-}" ] && [ -f "${CONFIG_FILE-}" ]; then
    . "${CONFIG_FILE}"
fi

PK_FILE=${PKI}/PK.cer
KEK_FILES=${PKI}/KEK.cer,${PKI}/MicCorKEKCA2011_2011-06-24.crt
DB_FILES=${PKI}/db.cer,${PKI}/MicWinProPCA2011_2011-10-19.crt,${PKI}/MicCorUEFCA2011_2011-06-27.crt

gsutil mb -l "${GCP_REGION}" "gs://${GCP_BUCKET}" || true
gsutil pap set enforced "gs://${GCP_BUCKET}" || true
gsutil cp "${GCP_IMAGE_PATH}" "gs://${GCP_BUCKET}/${GCP_IMAGE_FILENAME}"
gcloud compute images create "${GCP_IMAGE_NAME}" \
    "--family=${GCP_IMAGE_FAMILY}" \
    "--source-uri=gs://${GCP_BUCKET}/${GCP_IMAGE_FILENAME}" \
    "--guest-os-features=GVNIC,SEV_CAPABLE,VIRTIO_SCSI_MULTIQUEUE,UEFI_COMPATIBLE" \
    "--platform-key-file=${PK_FILE}" \
    "--key-exchange-key-file=${KEK_FILES}" \
    "--signature-database-file=${DB_FILES}" \
    "--project=${GCP_PROJECT}"
gcloud compute images add-iam-policy-binding "${GCP_IMAGE_NAME}" \
		"--project=${GCP_PROJECT}" \
		--member='allAuthenticatedUsers' \
		--role='roles/compute.imageUser'
gsutil rm "gs://${GCP_BUCKET}/${GCP_IMAGE_FILENAME}"
