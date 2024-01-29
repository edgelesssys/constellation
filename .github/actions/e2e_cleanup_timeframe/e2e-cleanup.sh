#!/bin/env bash

function get_e2e_test_ids_on_date {
  ids="$(gh run list --created "$1" --json createdAt,workflowName,databaseId --jq '.[] | select(.workflowName | contains("e2e")) | .databaseId' -R edgelesssys/constellation)"
  echo $ids
}

function download_tfstate_artifact {
  gh run download "$1" -p "terraform-state-*" -R edgelesssys/constellation &>/dev/null
}

function delete_resources {
  cd $1/constellation-terraform
  terraform destroy -auto-approve &>/dev/null
  cd ../../
  echo delete $1
}

function delete_iam_config {
  cd $1/constellation-iam-terraform
  terraform destroy -auto-approve &>/dev/null
  cd ../../
  echo delete iam $1
}

if [[ -z $1 ]]; then
  echo "No password for artifact decryption provided!"
  echo "Usage: ./e2e-cleanup.sh <ARTIFACT_DECRYPT_PASSWD>"
  exit 1
fi

artifact_pwd=$1

shopt -s nullglob

start_date=$(date "+%Y-%m-%d")
end_date=$(date --date "-7 day" "+%Y-%m-%d")
dates_to_clean=()

while [[ "$end_date" != "$start_date" ]]; do
  dates_to_clean+=($end_date)
  end_date=$(date --date "$end_date +1 day" "+%Y-%m-%d")
done

echo "[*] retrieving run IDs for cleanup"
database_ids=()
for d in ${dates_to_clean[*]}; do
  echo "    retrieving run IDs from $d"
  database_ids+=($(get_e2e_test_ids_on_date $d))
done

echo "[*] downloading terraform state artifacts"
for id in ${database_ids[*]}; do
  echo "    downloading from workflow $id"
  download_tfstate_artifact $id
done

echo "[*] extracting artifacts"
for artifact in ./terraform-state-*.zip; do
  echo "    extracting $artifact"

  mkdir ${artifact%.*}

  unzip "$artifact"
  unzip artifact.zip -d ${artifact%.*} -P "$artifact_pwd"

  rm "$artifact"
  rm artifact.zip
done

echo "[*] deleting resources"
for directory in ./terraform-state-*; do
  echo "    deleting resources in $directory"
  delete_resources $directory
  echo "    deleting IAM configuration in $directory"
  delete_iam_config $directory
done

exit 0
