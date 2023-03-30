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

# Go code conventions

## General

Adhere to the style and best practices described in [Effective Go](https://golang.org/doc/effective_go.html). Read [Common Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) for further information.

## Linting

This projects uses [golangci-lint](https://golangci-lint.run/) for linting.
You can [install golangci-lint](https://golangci-lint.run/usage/install/#linux-and-windows) locally,
but there is also a CI action to ensure compliance.

It is also recommended to use golangci-lint (and [gofumpt](https://github.com/mvdan/gofumpt) as formatter) in your IDE, by adding the recommended VS Code Settings or by [configuring it yourself](https://golangci-lint.run/usage/integrations/#editor-integration)

## Logging

We use a [custom subset](/internal/logger/) of [zap](https://pkg.go.dev/go.uber.org/zap) to provide logging for Constellation's services and components.
Usage instructions can be found in the package documentation.

Certain components may further specify a subset of the logger for their use. For example, the CLI has a debug-only logger, restricting the use of the logger to only `Debugf()`.

Further we try to adhere to the following guidelines:

* Do not log potentially sensitive information, e.g. variables that contain keys, secrets or otherwise protected information.

* Start log messages in uppercase and end without a punctuation mark. Exclamation, question marks, or ellipsis may be used where appropriate.

  Example:

  ```Go
  log.Infof("This is a log message")
  log.Infof("Waiting to do something...")
  log.Error("A critical error occurred!")
  ```

* Use the `With()` method to add structured context to your log messages. The context tags should be easily searchable to allow for easy log filtering. Try to keep consistent tag naming!

  Example:

  ```Go
  log.With(zap.Error(someError), zap.String("ip", "192.0.2.1")).Errorf("Connecting to IP failed")
  ```

* Log messages may use format strings to produce human readable messages. However, the information should also be present as structured context fields if it might be valuable for debugging purposes.

  Example:

  ```Go
  log.Infof("Starting server on %s:%s", addr, port)
  ```

* Usage of the `Fatalf()` method should be constrained to the main package of an application only!

* Use log levels to configure how detailed the logs of you application should be.

  * `Debugf()` for log low level and detailed information. This may include variable dumps, but should not disclose sensitive information, e.g. keys or secret tokens.
  * `Infof()` for general information.
  * `Warnf()` for information that may indicate unwanted behavior, but is not an application error. Commonly used by retry loops.
  * `Errorf()` to log information about any errors that occurred.
  * `Fatalf()` to log information about any errors that occurred and then exit the program. Should only be used in the main package of an application.

* Loggers passed to subpackages of an application may use the `Named()` method for better understanding of where a message originated.

  Example:

  ```Go
  grpcServer, err := server.New(log.Named("server"))
  ```

## Nested Go modules

As this project contains nested Go modules, we use a Go work file to ease integration with IDEs. You can find an introduction in the [Go workspace tutorial](https://go.dev/doc/tutorial/workspaces).

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

## Code documentation

Documentation of the latest release are available on [pkg.go.dev](https://pkg.go.dev/github.com/edgelesssys/constellation/v2).

Alternatively use `pkgsite` to host your own documentation server and view documentation for the local version of your code.

<details>
<summary>View installation instructions</summary>

```shell
CONSTELLATION_DIR=</path/to/your/local/report>
git clone https://github.com/golang/pkgsite && cd pkgsite
go install ./cmd/pkgsite
cd "${CONSTELLATION_DIR}
pkgsite
```

You can now view the documentation on http://localhost:8080/github.com/edgelesssys/constellation/v2
</details>

## PR conventions

Our changelog is generated from PR titles. Which PR is listed in which category is determined by labels, see the [release.yml](/.github/release.yml).

The PR title should be structured in one of the following ways:

```
<module>: <title>
```

Where the `<module>` is

- the top level directory of the microservice or component, e.g., `joinservice`, `disk-mapper`, `upgrade-agent` but also `docs` and `rfc`
- in internal, the second level directory
- `deps` for dependency upgrades
- `ci` for things that are CI related

and `<title>` is all lower case (except proper names, including acronyms).
Ticket numbers shouldn't be part of the title.

In case the scope of your PR is too wide, use the alternative format.

```
<Title>
```

and `<Title>` starts with a capital letter.

## CLI reference

The command reference within the CLI should follow the following conventions:

- Short description: Starts with a capital letter, beginnings of sentences, names and acronyms are capitalized, ends without a period. It should be a single sentence.
- Long description: Starts with a capital letter, beginnings of sentences, names and acronyms are capitalized, ends with a period.
  - If the long description contains multiple sentences, the first sentence is formatted as a long description, followed by 2 newlines and the rest of the sentences. The rest of the sentences start with a capital letter, beginnings of sentences, names and acronyms are capitalized and each sentence ends with a period.
- Flag: Starts with a lowercase letter, beginnings of sentences, names and acronyms are capitalized, ends without a period.
  - If a flag contains multiple sentences, the first sentence is formatted as a normal flag, followed by a newline and the rest of the sentences. The rest of the sentences start with a capital letter, beginnings of sentences, names and acronyms are capitalized and each sentence ends with a period.

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

# Shell script code conventions

We use [shellcheck](https://github.com/koalaman/shellcheck) to ensure code quality.
You might want to install an [IDE extension](https://marketplace.visualstudio.com/items?itemName=timonwong.shellcheck).
