# Reproducing Conformance Test Results

## Prerequisites

[Install & configure `gcloud` CLI](https://cloud.google.com/sdk/gcloud) for access to GCP.

[Install kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/) for working with Kubernetes

For more information [follow our documentation.](https://docs.edgeless.systems/constellation/latest/#/getting-started/install)

Additionally, [Sonobuoy CLI is required.](https://github.com/vmware-tanzu/sonobuoy/releases)
These tests results were produced using Sonobuoy v0.56.4.

## Provision Constellation Cluster

```sh
constellation create gcp 1 2 n2d-standard-4 -y
constellation init
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
./constellation terminate
rm constellation-mastersecret.base64
```

## Run CIS Benchmark Tests

```sh
# Runs for <1 min.
sonobuoy run --plugin https://raw.githubusercontent.com/vmware-tanzu/sonobuoy-plugins/master/cis-benchmarks/kube-bench-plugin.yaml --plugin https://raw.githubusercontent.com/vmware-tanzu/sonobuoy-plugins/master/cis-benchmarks/kube-bench-master-plugin.yaml --wait
# ... download & display results.
outfile=$(sonobuoy retrieve)
sonobuoy results $outfiles
```
