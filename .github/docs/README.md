# Actions & Workflows

## Manual Trigger (workflow_dispatch)

It is currently not possible to run a `workflow_dispatch` based workflow on a specific branch from the WebUI. If you need to do this, use the [GitHub CLI](https://github.com/cli/cli):

```bash
gh workflow run e2e-test.yml \
    --ref feat/e2e_pipeline \                       # On your specific branch!
    -F autoscale=false -F cloudProvider=gcp \       # With your ...
    -F controlNodesCount=1 -F workerNodesCount=2 \  # ... settings
    -F machineType=n2d-standard-2
```

### E2E Test Suites

Here are some examples for test suits you might want to run. Values for `sonobuoyTestSuiteCmd`:

* `--mode quick`
    * Runs a set of tests that are known to be quick to execute!
* `--e2e-focus "Services should be able to create a functioning NodePort service"`
    * Runs a specific test
* `--mode certified-conformance`
    * For K8s conformance certification test suite

Check [Sonobuoy docs](https://sonobuoy.io/docs/latest/e2eplugin/) for more examples.

## Local Development

Using [nektos/act](https://github.com/nektos/act) you can run GitHub actions locally.

### Specific Jobs

```bash
act -j e2e-test
```

### Wireguard

When running actions that use Wireguard, you need to provide additional capabilities to Docker:

```bash
act --secret-file secrets.env --container-cap-add NET_ADMIN --container-cap-add SYS_MODULE --privileged
```
### Authorizing GCP

For creating Kubernetes clusters in GCP a local copy of the service account secret is required.

1. [Create a new service account key](https://console.cloud.google.com/iam-admin/serviceaccounts/details/112741463528383500960/keys?authuser=0&project=constellation-331613&supportedpurview=project)
2. Create a compact (one line) JSON representation of the file `jq -c`
3. Create a secrets file for act to consume:

```bash
$ cat secrets.env
GCP_SERVICE_ACCOUNT={"type":"service_account", ... }

$ act --secret-file secrets.env
```
