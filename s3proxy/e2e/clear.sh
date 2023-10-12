#!/usr/bin/env bash

# This script can be used to remove buckets from S3.
# It will empty the buckets and eventually remove them.
# It is expected that the script throws some errors. E.g. "Bucket is missing Object Lock Configuration" or "Invalid type for parameter Delete.Objects, value: None [..]"
# These can be ignored. The first error is thrown if the bucket does not have object lock enabled. The second error is thrown if the bucket is already empty.
# In both cases the bucket is empty and can be removed.

# Usage: ./clear.sh <prefix>
# The prefix is necessary, as otherwise all buckets are removed.

readonly prefix=$1

if [ -z "$prefix" ]; then
  echo "Usage: $0 <prefix>"
  echo "WARNING: If you don't provide a prefix, all buckets are destroyed."
  exit 1
fi

restore_aws_page="$AWS_PAGER"
export AWS_PAGER=""

function empty_bucket() {
  # List all object versions in the bucket
  versions=$(aws s3api list-object-versions --bucket "$1" --output=json --query='{Objects: Versions[].{Key:Key,VersionId:VersionId}}')

  # Remove all legal holds
  for version in "$versions"; do
    key=$(echo "$version" | jq '.Objects[0].Key' | tr -d '"')
    aws s3api put-object-legal-hold --bucket "$1" --key "$key" --legal-hold Status=OFF
  done
  # Delete all object versions
  aws s3api delete-objects --bucket "$1" --delete "$versions" || true

  # List all delete markers in the bucket
  markers=$(aws s3api list-object-versions --bucket "$1" --output=json --query='{Objects: DeleteMarkers[].{Key:Key,VersionId:VersionId}}')

  # Remove all delete markers
  aws s3api delete-objects --bucket "$1" --delete "$markers" || true
}

for i in $(aws s3api list-buckets --query "Buckets[?starts_with(Name, '${prefix}')].Name" --output text); do
  empty_bucket $i
  aws s3 rb s3://$i
done

export AWS_PAGER="$restore_aws_page"
