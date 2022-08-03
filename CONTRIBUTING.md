## Testing

Run all unit tests locally with

```sh
cd build
cmake ..
ctest
```

[Run CI e2e tests](/.github/docs/README.md)

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
