# Writing to customers: style policy

Whenever you write text facing the customer (e.g, docs, warnings, errors), follow the [Microsoft Style Guide](https://learn.microsoft.com/en-us/style-guide/welcome/).
For quick reference, check [Top 10 tips for Microsoft style and voice](https://learn.microsoft.com/en-us/style-guide/top-10-tips-style-voice).

# Go code conventions

## General

Adhere to the style and best practices described in [Effective Go](https://golang.org/doc/effective_go.html). Read [Common Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) for further information.

This project also aims to follow the [Go Proverbs](https://go-proverbs.github.io/).

## Linting

This projects uses [golangci-lint](https://golangci-lint.run/) for linting.
You can [install golangci-lint](https://golangci-lint.run/usage/install/#linux-and-windows) locally,
but there is also a CI action to ensure compliance.

It is also recommended to use golangci-lint (and [gofumpt](https://github.com/mvdan/gofumpt) as formatter) in your IDE, by adding the recommended VS Code Settings or by [configuring it yourself](https://golangci-lint.run/usage/integrations/#editor-integration)

## Logging

We use [slog](https://pkg.go.dev/log/slog) for logging.
Usage instructions can be found in the package documentation.

Certain components may further specify a subset of the logger for their use. For example, the CLI has a debug-only logger, restricting the use of the logger to only `Debug()`.

Further we try to adhere to the following guidelines:

* Do not log potentially sensitive information, e.g. variables that contain keys, secrets or otherwise protected information.

* Start log messages in uppercase and end without a punctuation mark. Exclamation, question marks, or ellipsis may be used where appropriate.

  Example:

  ```Go
  log.Info("This is a log message")
  log.Info("Waiting to do something...")
  log.Error("A critical error occurred!")
  ```

* Use the `With()` method to add structured context to your log messages. The context tags should be easily searchable to allow for easy log filtering. Try to keep consistent tag naming!

  Example:

  ```Go
  log.With(slog.Any("error" someError), slog.String("ip", "192.0.2.1")).Error("Connecting to IP failed")
  ```

* Log messages may use format strings to produce human readable messages. However, the information should also be present as structured context fields if it might be valuable for debugging purposes. So, instead of writing

  ```Go
  log.Info(fmt.Sprintf("Starting server on %s:%s", addr, port))
  ```

  You should write

  ```Go
  log.With(slog.String("address", addr), slog.Int("port", port)).Info("Starting server")
  ```

* Use log levels to configure how detailed the logs of you application should be.

  * `Debug()` for log low level and detailed information. This may include variable dumps, but should not disclose sensitive information, e.g. keys or secret tokens.
  * `Info()` for general information.
  * `Warn()` for information that may indicate unwanted behavior, but is not an application error. Commonly used by retry loops.
  * `Error()` to log information about any errors that occurred.

* Loggers passed to subpackages of an application may use the `WithGroup()` method for better understanding of where a message originated.

  Example:

  ```Go
  grpcServer, err := server.New(log.WithGroup("server"))
  ```

## Nested Go modules

As this project contains nested Go modules, we use a Go work file to ease integration with IDEs. You can find an introduction in the [Go workspace tutorial](https://go.dev/doc/tutorial/workspaces).

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

You can now view the documentation on <http://localhost:8080/github.com/edgelesssys/constellation/v2>
</details>

## Adding to a package

Adding new functionality to a package is often required whenever new features for Constellation are implemented.

Before adding to a package, ask yourself:

* Does this feature implement functionality outside the scope of this package?
  * Keep in mind the design goals of the package you are proposing to edit
  * If yes, consider adding it to a different package, or [add a new one](#adding-new-packages)
* Does another package already provide this functionality?
  * If yes, use that package instead
* Do other parts of Constellation (binaries, tools, etc.) also require this feature?
  * If yes, consider adding it an existing, or create a new package, in the global [`internal`](../internal/) package instead.
* Do other parts of Constellation already implement this feature?
  * If yes, evaluate if you can reasonably move the functionality from that part of Constellation to the global [`internal`](../internal/) package to avoid code duplication.

If the answer to all of the questions was "No", extend the package with your chosen functionality.

Remember to:

* Update the package description if the package received any major changes
* Add unit tests

## Adding new packages

If you are implementing a feature you might find yourself adding code that does not share anything with the existing packages of the binary/tool you are working on.
In that case you might need to add a new package.

Before adding a new package, ask yourself:

* Does your new package provide functionality that could reasonably be added to one of the existing packages?
  * Keep in mind the design goals of the existing package: Don't add out of scope functionality to an existing package to avoid creating a new one.
* Do other parts of Constellation (binaries, tools, etc.) also require this feature?
  * If yes, consider adding the new package to the global [`internal`](../internal/) package.
* Do other parts of Constellation already implement this feature?
  * If yes, evaluate if you can reasonably move the functionality from that part of Constellation to the global [`internal`](../internal/) package to avoid code duplication.

If the answer to all of the questions was "No', add the new package to the binary/tool you are working on.

Remember to:

* Add a description to your new package
* Add unit tests

## CLI reference

The command reference within the CLI should follow the following conventions:

* Short description: Starts with a capital letter, beginnings of sentences, names and acronyms are capitalized, ends without a period. It should be a single sentence.
* Long description: Starts with a capital letter, beginnings of sentences, names and acronyms are capitalized, ends with a period.
  * If the long description contains multiple sentences, the first sentence is formatted as a long description, followed by 2 newlines and the rest of the sentences. The rest of the sentences start with a capital letter, beginnings of sentences, names and acronyms are capitalized and each sentence ends with a period.
* Flag: Starts with a lowercase letter, beginnings of sentences, names and acronyms are capitalized, ends without a period.
  * If a flag contains multiple sentences, the first sentence is formatted as a normal flag, followed by a newline and the rest of the sentences. The rest of the sentences start with a capital letter, beginnings of sentences, names and acronyms are capitalized and each sentence ends with a period.

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
