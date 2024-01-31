#!/bin/env bash

# get_e2e_test_ids_on_date gets all workflow IDs of workflows that contain "e2e" on a specific date.
function get_e2e_test_ids_on_date {
  ids="$(gh run list --created "$1" --json createdAt,workflowName,databaseId --jq '.[] | select(.workflowName | contains("e2e")) | .databaseId' -R edgelesssys/constellation)"
  echo $ids
}

# download_tfstate_artifact downloads all artifacts matching the pattern terraform-state-* from a given workflow ID.
function download_tfstate_artifact {
  gh run download "$1" -p "terraform-state-*" -R edgelesssys/constellation > /dev/null
}

# delete_resources runs terraform destroy on the constellation-terraform subfolder of a given folder.
function delete_resources {
  cd $1/constellation-terraform
  terraform init > /dev/null # first, install plugins
  terraform destroy -auto-approve > /dev/null
  cd ../../
}

# delete_iam_config runs terraform destroy on the constellation-iam-terraform subfolder of a given folder.
function delete_iam_config {
  cd $1/constellation-iam-terraform
  terraform init > /dev/null # first, install plugins
  terraform destroy -auto-approve > /dev/null
  cd ../../
}

# check if the password for artifact decryption was given
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

# get all dates of the last week
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
for directory in ./terraform-state-*; do
  echo "    extracting $directory"

  # extract and decrypt the artifact
  unzip -d ${directory} -P "$artifact_pwd" $directory/archive.zip > /dev/null
done

echo "[*] deleting resources"
for directory in ./terraform-state-*; do
  echo "    deleting resources in $directory"
  delete_resources $directory
  echo "    deleting IAM configuration in $directory"
  delete_iam_config $directory
done

exit 0
