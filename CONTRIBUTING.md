## Testing

Run all unit tests locally with

```sh
cd build
cmake ..
ctest
```

### E2E Test

Requirement: Kernel WireGuard, Docker
```sh
docker build -f Dockerfile.e2e -t constellation-e2e .
```
For the AWS test run
```sh
docker run -it --cap-add=NET_ADMIN --env GITHUB_TOKEN="$(cat ~/.netrc)" --env BRANCH="main" --env aws_access_key_id=XXX --env aws_secret_access_key=XXX constellation-e2e /initiateAWS.sh
```
For the gcp test run
```sh
docker run -it --cap-add=NET_ADMIN --env GITHUB_TOKEN="$(cat ~/.netrc)" --env BRANCH="main" --env GCLOUD_CREDENTIALS="$(cat ./constellation-keyfile.json)" constellation-e2e /initiategcloud.sh
```

## Linting

This projects uses [golangci-lint](https://golangci-lint.run/) for linting.
You can [install golangci-lint](https://golangci-lint.run/usage/install/#linux-and-windows) locally,
but there is also a CI action to ensure compliance.

To locally run all configured linters, execute

```
golangci-lint run ./...
```

It is also recommended to use golangci-lint (and [gofumpt](https://github.com/mvdan/gofumpt) as formatter) in your IDE, by adding the recommended VS Code Settings or by [configuring it yourself](https://golangci-lint.run/usage/integrations/#editor-integration)


## Recommended VS Code Settings

The following can be added to your personal `settings.json`, but it is recommended to add it to
the `<REPOSITORY>/.vscode/settings.json` repo, so the settings will only affect this repository.

```jsonc
    // Use gofumpt as formatter.
    "gopls": {
      "formatting.gofumpt": true,
    },
    // Use golangci-lint as linter. Make sure you've installed it.
    "go.lintTool":"golangci-lint",
    "go.lintFlags": ["--fast"],
    // You can easily show Go test coverage by running a package test.
    "go.coverageOptions": "showUncoveredCodeOnly",
    // Executing unit tests with race detection.
    // You can add preferences like "-v" or "-count=1"
    "go.testFlags": ["-race"],
    // Enable language features for files with build tags.
    // Attention! This leads to integration test being executed when
    // running a package test within a package containing integration
    // tests.
    "go.buildTags": "integration",
```
