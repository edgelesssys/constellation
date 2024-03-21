#!/bin/bash

# GitHub CLI api
# https://cli.github.com/manual/gh_api

# gh api \
#   --method DELETE \
#   -H "Accept: application/vnd.github+json" \
#   -H "X-GitHub-Api-Version: 2022-11-28" \
#   /repos/OWNER/REPO/actions/artifacts/ARTIFACT_ID

workflow_id=$1
artifact_name=$2

if [[ -n workflow_id ]]; then;
  echo "[X] No workflow id provided."
  echo "Usage: delete_artifact.sh <WORKFLOW_ID> <ARTIFACT_NAME>"
  exit 1
fi

if [[ -n artifact_name ]]; then;
  echo "[X] No artifact name provided."
  echo "Usage: delete_artifact.sh <WORKFLOW_ID> <ARTIFACT_NAME>"
  exit 1
fi
