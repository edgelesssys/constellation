# VS Code

## Recommended Settings

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
    // Attention! This leads to integration/e2e tests being executed when
    // running a package test within a package containing integration/e2e
    // tests.
    "go.buildTags": "integration e2e",
```

For some inexplicable reason, the `"go.lintTool":"golangci-lint",` might be overwritten. In case you don't get all linter suggestions, you might want to check the value of `go.lintTool` in the UI settings and make sure it is also set to `golangci-lint`.

Additionally, we use the [Redhat YAML formatter](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml) to have uniform formatting in our `.yaml` files.

## Recommended extensions

* Bazel (BazelBuild.vscode-bazel): Bazel syntax highlighting and more
* Go (golang.go): Go language support for VS Code
* HashiCorp Terraform (hashicorp.terraform): Syntax highlighting for Terraform files
* ShellCheck (timonwong.shellcheck): Shell script linter
* vscode-proto3 (zxh404.vscode-proto3): Protobuf language support
* Code Spell Checker (streetsidesoftware.code-spell-checker): Highlights potential spelling mistakes
* Helm Intellisense: (Tim-Koehler.helm-intellisense): Syntax highlighting and more for Helm charts (not available on [Open VSX Registry](https://open-vsx.org/))
* YAML (redhat.vscode-yaml): YAML language support. (Does not work with Helm charts)
* markdownlint (DavidAnson.vscode-markdownlint): Markdown linter

## Bazel support

You might also consider to set up Bazel in the IDE (see [here](./bazel.md#vs-code-integration)).
