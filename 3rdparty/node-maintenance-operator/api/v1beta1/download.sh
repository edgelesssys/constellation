#!/usr/bin/env bash

set -exuo pipefail
shopt -s inherit_errexit

# this script is used to download the CRD API definition from the node-maintenance-operator.

archive_url=$1
prefix=$2
expected_sha256=$3
curl -fsSL -o download.tar.gz "${archive_url}"
echo "${expected_sha256} download.tar.gz"
sha256sum download.tar.gz
echo "${expected_sha256} download.tar.gz" | sha256sum -c -
tar --strip-components=3 -xzf download.tar.gz "${prefix}"
rm \
  nodemaintenance_webhook.go \
  nodemaintenance_webhook_test.go \
  webhook_suite_test.go
rm download.tar.gz
