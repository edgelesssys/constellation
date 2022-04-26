# Reproducing Conformance Test Results

## Prerequisites

[Install & configure `gcloud` CLI](https://cloud.google.com/sdk/gcloud) for access to GCP.

[Install WireGuard](https://www.wireguard.com/install/) for connecting to your cluster

[Install kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/) for working with Kubernetes

For more information [follow our documentation.](https://constellation-docs.edgeless.systems/6c320851-bdd2-41d5-bf10-e27427398692/#/getting-started/install)

Additionally, [Sonobuoy CLI is required.](https://github.com/vmware-tanzu/sonobuoy/releases)
These tests results were produced using Sonobuoy v0.56.4.

## Provision Constellation Cluster

```sh
constellation create gcp 1 2 n2d-standard-2 -y
constellation init
wg-quick up ./wg0.conf
export KUBECONFIG="$PWD/constellation-admin.conf"
```

## Run Conformance Tests

```sh
# Runs for ~2 hours.
sonobuoy run --mode certified-conformance
# Once status shows tests have completed...
sonobuoy status
# ... download & display results.
outfile=$(sonobuoy retrieve)
sonobuoy results $outfile
```

## Fetch Test Log & Report

The provided `e2e.log` & `junit_01.xml` were fetched like this:

```sh
tar -xvf $outfile
cat plugins/e2e/results/global/e2e.log
cat plugins/e2e/results/global/junit_01.xml
```

## Cleanup

```sh
# Remove test deployments
sonobuoy delete --wait
# Or, shutdown cluster
wg-quick down ./wg0.conf
./constellation terminate
rm constellation-mastersecret.base64
```
