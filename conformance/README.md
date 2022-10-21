# Reproducing Conformance Test Results

## Prerequisites

[Follow the docs on how to set up cloud credentials for GCP](https://docs.edgeless.systems/constellation/getting-started/install#set-up-cloud-credentials)

[Install kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/) for working with Kubernetes

Additionally, [Sonobuoy CLI is required.](https://github.com/vmware-tanzu/sonobuoy/releases)
These tests results were produced using Sonobuoy v0.56.10

## Provision Constellation Cluster

```sh
constellation config generate gcp
```

Fill the config with the needed values for your cloud subscription.
Set `kubernetesVersion: "1.x"`.

```sh
constellation config fetch-measurements
constellation create -c3 -w2 -y
constellation init --conformance
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
constellation terminate
rm constellation-mastersecret.base64
```
