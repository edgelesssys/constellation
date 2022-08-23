# Actions & Workflows

## Manual Trigger (workflow_dispatch)

It is currently not possible to run a `workflow_dispatch` based workflow on a specific branch, while it is not yet available in `main` branch, from the WebUI. If you would like to test your pipeline changes on a branch, use the [GitHub CLI](https://github.com/cli/cli):

```bash
gh workflow run e2e-test-manual.yml \
    --ref feat/e2e_pipeline \                       # On your specific branch!
    -F autoscale=false -F cloudProvider=gcp \       # With your ...
    -F controlNodesCount=1 -F workerNodesCount=2 \  # ... settings
    -F machineType=n2d-standard-2
```

### E2E Test Suites

Here are some examples for test suits you might want to run. Values for `sonobuoyTestSuiteCmd`:

* `--mode quick`
    * Runs a set of tests that are known to be quick to execute! (<1 min)
* `--e2e-focus "Services should be able to create a functioning NodePort service"`
    * Runs a specific test
* `--mode certified-conformance`
    * For K8s conformance certification test suite

Check [Sonobuoy docs](https://sonobuoy.io/docs/latest/e2eplugin/) for more examples.

When using `--mode` be aware that `--e2e-focus` and `e2e-skip` will be overwritten. [Check in the source code](https://github.com/vmware-tanzu/sonobuoy/blob/e709787426316423a4821927b1749d5bcc90cb8c/cmd/sonobuoy/app/modes.go#L130) what the different modes do.

## Local Development

Using [nektos/act](https://github.com/nektos/act) you can run GitHub actions locally.

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
      "autoscale": false,
      "cloudProvider": "gcp",
      "machineType": "n2d-standard-2",
      "sonobuoyTestSuiteCmd": "--mode quick"
  }
}
```

Then run act with the event as input:

```bash
act -j e2e-test-manual --eventpath event.json
```

### Authorizing GCP

For creating Kubernetes clusters in GCP a local copy of the service account secret is required.

1. [Create a new service account key](https://console.cloud.google.com/iam-admin/serviceaccounts/details/112741463528383500960/keys?authuser=0&project=constellation-331613&supportedpurview=project)
2. Create a compact (one line) JSON representation of the file `jq -c`
3. Store in [GitHub Action Secret](https://github.com/edgelesssys/constellation/settings/secrets/actions) or create a local secret file for act to consume:

```bash
$ cat secrets.env
GCP_SERVICE_ACCOUNT={"type":"service_account", ... }

$ act --secret-file secrets.env
```

### Authorizing Azure

Create a new service principal:

```bash
az ad sp create-for-rbac --name "github-actions-e2e-tests" --role contributor --scopes /subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435 --sdk-auth
az role assignment create --role "User Access Administrator" --scope /subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435 --assignee <SERVICE_PRINCIPAL_CLIENT_ID>
```

Next, [add API permissions to Managed Identity](https://github.com/edgelesssys/wiki/blob/master/other_tech/azure.md#adding-api-permission-to-managed-identity)

Store output of `az ad sp ...` in [GitHub Action Secret](https://github.com/edgelesssys/constellation/settings/secrets/actions) or create a local secret file for act to consume.

## Image versions

The [build-coreos](../workflows/build-coreos.yml) workflow can be used to trigger an image build.

The workflow can be used to build debug or release images.
A debug image uses [`debugd`](../../debugd/) as its bootstrapper binary, while release images use the actual [`bootstrapper`](../../bootstrapper/)
Workflows for the main branch will always build debug images.

The image will be named and categorized depending on the branch the build is triggered from.
In the following, __Release__ refers to non debug images build from a release branch, e.g. `release/v1.4.0`,
__Debug__ refers to debug images build from either main or a release branch,
and __Branch__ refers to any image build from a branch that is not main or a release branch.
Non debug images built from main follow the __Branch__ image naming scheme.

### GCP

Type | Image Family | Image Name
-|-|-
Release | constellation | constellation-v\<major\>-\<minor\>-\<patch\>
Debug | constellation-debug-v\<major\>-\<minor\>-\<patch\> | constellation-\<commit-timestamp\>
Branch | constellation-\<branch-name\> | constellation-\<commit-timestamp\>

Example:
Type | Image Family | Image Name | List command
-|-|-|-
Release | constellation | constellation-v1-5-0 | `gcloud compute images list --filter="family~'^constellation$'" --sort-by=creationTimestamp --project constellation-images --uri \| sed 's#https://www.googleapis.com/compute/v1/##'`
Debug | constellation-debug-v1-5-0 | constellation-20220912123456 | `gcloud compute images list --filter="family~'constellation-debug-v.+'" --sort-by=creationTimestamp --project constellation-images --uri \| sed 's#https://www.googleapis.com/compute/v1/##'`
Branch | constellation-ref-cli | constellation-20220912123456 | `gcloud compute images list --filter="family~'constellation-$(go run $(git rev-parse --show-toplevel)/hack/pseudo-version/pseudo-version.go -print-branch)'" --sort-by=creationTimestamp --project constellation-images --uri \| sed 's#https://www.googleapis.com/compute/v1/##'`

### Azure

Type | Gallery | Image Definition | Image Version
-|-|-|-
Release | Constellation | constellation | \<major\>.\<minor\>.\<patch\>
Debug | Constellation_Debug | v\<major\>.\<minor\>.\<patch\> | \<commit-timestamp\>
Branch | Constellation_Testing | \<branch-name\> | \<commit-timestamp\>

Example:

Type | Gallery | Image Definition | Image Version | List command
-|-|-|-|-
Release | Constellation | constellation | 1.5.0 | `az sig image-version list --resource-group constellation-images --gallery-name Constellation --gallery-image-definition constellation --query "sort_by([], &publishingProfile.publishedDate)[].id" -o table`
Debug | Constellation_Debug | v1.5.0 | 2022.0912.123456 | `az sig image-version list --resource-group constellation-images --gallery-name Constellation_Debug --gallery-image-definition v1.5.0 --query "sort_by([], &publishingProfile.publishedDate)[].id" -o table`
Branch | Constellation_Testing | ref-cli | 2022.0912.123456 | `az sig image-version list --resource-group constellation-images --gallery-name Constellation_Testing --gallery-image-definition $(go run $(git rev-parse --show-toplevel)/hack/pseudo-version/pseudo-version.go -print-branch) --query "sort_by([], &publishingProfile.publishedDate)[].id" -o table`
