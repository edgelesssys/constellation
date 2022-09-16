# Process conventions

## Pull request process

Submissions should remain focused in scope and avoid containing unrelated commits.
For pull requests, we employ the following workflow:

1. Fork the repository to your own GitHub account
2. Create a branch locally with a descriptive name
3. Commit changes to the branch
4. Write your code according to our development guidelines
5. Push changes to your fork
6. Clean up your commit history
7. Open a PR in our repository and summarize the changes in the description

## Reporting issues and bugs, asking questions

This project uses the GitHub issue tracker. Please check the existing issues before submitting to avoid duplicates.

To report a security issue, contact security@edgeless.systems.

Your bug report should cover the following points:

* A quick summary and/or background of the issue
* Steps to reproduce (be specific, e.g., provide sample code)
* What you expected would happen
* What actually happens
* Further notes:
  * Thoughts on possible causes
  * Tested workarounds or fixes

## Major changes and feature requests

You should discuss larger changes and feature requests with the maintainers. Please open an issue describing your plans.

[Run CI e2e tests](/.github/docs/README.md)

# Code conventions

## General

Adhere to the style and best practices described in [Effective Go](https://golang.org/doc/effective_go.html). Read [Common Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) for further information.

## Linting

This projects uses [golangci-lint](https://golangci-lint.run/) for linting.
You can [install golangci-lint](https://golangci-lint.run/usage/install/#linux-and-windows) locally,
but there is also a CI action to ensure compliance.

To locally run all configured linters, execute

```sh
golangci-lint run ./...
```

It is also recommended to use golangci-lint (and [gofumpt](https://github.com/mvdan/gofumpt) as formatter) in your IDE, by adding the recommended VS Code Settings or by [configuring it yourself](https://golangci-lint.run/usage/integrations/#editor-integration)

## Nested Go modules

As this project contains nested Go modules, it is recommended to create a local Go workspace, so your IDE can lint multiple modules at once.

```go
go 1.18

use (
 .
 ./hack
 ./operators/constellation-node-operator
)
```

You can find an introduction in the [Go workspace tutorial](https://go.dev/doc/tutorial/workspaces).

If you have changed dependencies within a module and have run `go mod tidy`, you can use `go work sync` to sync versions of the same dependency of the different modules.

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

Additionally, we use the [Redhat YAML formatter](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml) to have uniform formatting in our `.yaml` files.

## Naming convention

### Network

IP addresses:

* ip: numeric IP address
* host: either IP address or hostname
* endpoint: host+port

### Keys

* key: symmetric key
* pubKey: public key
* privKey: private key
