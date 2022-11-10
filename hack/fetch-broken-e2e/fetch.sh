#!/usr/bin/env bash

set -euo pipefail
shopt -s inherit_errexit

LATEST_AZURE_RUNS=$(
  gh run list \
    -R edgelesssys/constellation \
    -w 'e2e Test Azure' \
    --json databaseId \
    -q '.[].databaseId'
)
echo "${LATEST_AZURE_RUNS}"
for RUN_ID in ${LATEST_AZURE_RUNS}; do
  # Might fail, because no state was written, because e2e pipeline failed early
  # Or, because state was downloaded by earlier run of this script
  gh run download "${RUN_ID}" \
    -R edgelesssys/constellation \
    -n constellation-state.json \
    -D azure/"${RUN_ID}" || true
done

LATEST_GCP_RUNS=$(
  gh run list \
    -R edgelesssys/constellation \
    -w 'e2e Test GCP' \
    --json databaseId \
    -q '.[].databaseId'
)
echo "${LATEST_GCP_RUNS}"
for RUN_ID in ${LATEST_GCP_RUNS}; do
  # Might fail, because no state was written, because e2e pipeline failed early
  # Or, because state was downloaded by earlier run of this script
  gh run download "${RUN_ID}" \
    -R edgelesssys/constellation \
    -n constellation-state.json \
    -D gcp/"${RUN_ID}" || true
done
