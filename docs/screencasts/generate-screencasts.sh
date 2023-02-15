#!/usr/bin/env bash
#
# This script prepares the environment for expect scripts to be recorded in,
# executes all scripts, and copies the .cast files to our doc's asset folder.
#
# Note: A cluster is created in GCP. Therefore you are expected to be logged in
# via `gcloud` CLI. You credentials at $HOME/.config/gcloud are mounted into the
# screenrecordings container. A full script run takes ~20min.
#

docker build -t screenrecodings docker

# Verify CLI
docker run -it -v "$(pwd)"/recordings:/recordings screenrecodings /scripts/verify-cli.expect
cp recordings/verify-cli.cast ../static/assets/verify-cli.cast

# Check SBOM
docker run -it -v "$(pwd)"/recordings:/recordings screenrecodings /scripts/check-sbom.expect
cp recordings/check-sbom.cast ../static/assets/check-sbom.cast

#
# Do not change the order of the following sections. A cluster is created,
# modified and finally destroyed. Otherwise resources might not get cleaned up.
#
# To get multiple recordings a dedicated script is used for each step.
# The Constellation working directory is shared via /constellation container folder.
#

# Create config
docker run -it \
    -v "${HOME}"/.config/gcloud:/root/.config/gcloud \
    -v "$(pwd)"/recordings:/recordings \
    -v "$(pwd)"/constellation:/constellation \
    screenrecodings /scripts/configure-cluster.expect
cp recordings/configure-cluster.cast ../static/assets/configure-cluster.cast

# Create cluster
docker run -it \
    -v "${HOME}"/.config/gcloud:/root/.config/gcloud \
    -v "$(pwd)"/recordings:/recordings \
    -v "$(pwd)"/constellation:/constellation \
    screenrecodings /scripts/create-cluster.expect
cp recordings/create-cluster.cast ../static/assets/create-cluster.cast

# Terminate cluster
docker run -it \
    -v "${HOME}"/.config/gcloud:/root/.config/gcloud \
    -v "$(pwd)"/recordings:/recordings \
    -v "$(pwd)"/constellation:/constellation \
    screenrecodings /scripts/terminate-cluster.expect
cp recordings/terminate-cluster.cast ../static/assets/terminate-cluster.cast

# Delete IAM
docker run -it \
    -v "${HOME}"/.config/gcloud:/root/.config/gcloud \
    -v "$(pwd)"/recordings:/recordings \
    -v "$(pwd)"/constellation:/constellation \
    screenrecodings /scripts/delete-iam.expect
cp recordings/delete-iam.cast ../static/assets/delete-iam.cast
