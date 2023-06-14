# Actions & Workflows

## Manual Trigger (workflow_dispatch)

It is currently not possible to run a `workflow_dispatch` based workflow on a specific branch, while it is not yet available in `main` branch, from the WebUI. If you would like to test your pipeline changes on a branch, use the [GitHub CLI](https://github.com/cli/cli):

```bash
gh workflow run e2e-test-manual.yml \
    --ref feat/e2e_pipeline \                       # On your specific branch!
    -F cloudProvider=gcp \       # With your ...
    -F controlNodesCount=1 -F workerNodesCount=2 \  # ... settings
    -F machineType=n2d-standard-4 \
    -F test=nop
```

### E2E Test Suites

Here are some examples for test suites you might want to run. Values for `sonobuoyTestSuiteCmd`:

* `--mode quick`
  * Runs a set of tests that are known to be quick to execute! (<1 min)
* `--e2e-focus "Services should be able to create a functioning NodePort service"`
  * Runs a specific test
* `--mode certified-conformance`
  * For K8s conformance certification test suite

Check [Sonobuoy docs](https://sonobuoy.io/docs/latest/e2eplugin/) for more examples.

When using `--mode` be aware that `--e2e-focus` and `e2e-skip` will be overwritten. [Check in the source code](https://github.com/vmware-tanzu/sonobuoy/blob/e709787426316423a4821927b1749d5bcc90cb8c/cmd/sonobuoy/app/modes.go#L130) what the different modes do.

## Local Development

Using [***act***](https://github.com/nektos/act) you can run GitHub actions locally.

**These instructions are for internal use.**
In case you want to use the E2E actions externally, you need to adjust other configuration parameters.
Check the assignments made in the [/.github/actions/e2e_test/action.yml](E2E action) and adjust any hard-coded values.

### Specific Jobs

```bash
act -j e2e-test-gcp
```

### Simulate a `workflow_dispatch` event

Create a new JSON file to describe the event ([relevant issue](https://github.com/nektos/act/issues/332), there are [no further information about structure of this file](https://github.com/nektos/act/blob/master/pkg/model/github_context.go#L11)):

```json
{
  "action": "workflow_dispatch",
  "inputs": {
      "workerNodesCount": "2",
      "controlNodesCount": "1",
      "cloudProvider": "gcp",
      "machineType": "n2d-standard-4",
      "sonobuoyTestSuiteCmd": "--mode quick"
  }
}
```

Then run *act* with the event as input:

```bash
act -j e2e-test-manual --eventpath event.json
```

### Authorizing GCP

For creating Kubernetes clusters in GCP a local copy of the service account secret is required.

1. [Create a new service account key](https://console.cloud.google.com/iam-admin/serviceaccounts/details/112741463528383500960/keys?authuser=0&project=constellation-331613&supportedpurview=project)
2. Create a compact (one line) JSON representation of the file `jq -c`
3. Store in a GitHub Action Secret called `GCP_SERVICE_ACCOUNT` or create a local secret file for *act* to consume:

```bash
$ cat secrets.env
GCP_SERVICE_ACCOUNT={"type":"service_account", ... }

$ act --secret-file secrets.env
```

In addition, you need to create a Service Account which Constellation itself is supposed to use. Refer to [First steps](https://docs.edgeless.systems/constellation/getting-started/first-steps#create-a-cluster) in the documentation on how to create it. What you need here specifically is the `gcpServiceAccountKey`, which needs to be stored in a secret called `GCP_CLUSTER_SERVICE_ACCOUNT`.

### Authorizing Azure

See [here](https://docs.edgeless.systems/constellation/workflows/config#creating-iam-credentials).
