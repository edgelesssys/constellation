# Bazel

Bazel is the primary build system for this project. It is used to build all Go code and will be used to build all artifacts in the future.
Still, we aim to keep the codebase compatible with `go build` and `go test` as well.
Whenever Go code is changed, you will have to run `bazel run //:tidy` to regenerate the Bazel build files for Go code.
Additionally, you need to update `MODULE.bazel`, together with `MODULE.bazel.lock`:

```
# if the steps below fail, try to recreate the lockfile from scratch by deleting it
bazel mod deps --lockfile_mode=update
bazel mod tidy
```

## Bazel commands

### Build

#### Useful defaults

```sh
bazel build //bootstrapper/cmd/bootstrapper:bootstrapper # build bootstrapper
bazel build //cli:cli_oss # build CLI
bazel build //cli:cli_oss_linux_amd64 # cross compile CLI for linux amd64
bazel build //cli:cli_oss_linux_arm64 # cross compile CLI for linux arm64
bazel build //cli:cli_oss_darwin_amd64 # cross compile CLI for mac amd64
bazel build //cli:cli_oss_darwin_arm64 # cross compile CLI for mac arm64
```

#### General

* `bazel build //...` - build all targets (when `.bazeloverwriterc` is specified see [here](./build-develop-deploy.md#settings))
* `bazel build //subfolder/...` - build all targets in a subfolder (recursive)
* `bazel build //subfolder:all` - build all targets in a subfolder (non-recursive)
* `bazel build //subfolder:target` - build single target

### Run

* `bazel run --run_under="cd $PWD &&" //cli:cli_oss -- create --yes` - build + run a target with arguments in current working directory

### Pre-PR checks

* `bazel test //...` - run all tests
* `bazel run //:generate` - execute code generation
* `bazel run //:tidy` - tidy, format and generate
* `bazel run //:check` - execute checks and linters. To reduce verbosity of non-critical output, you can set `SILENT=1 bazel run //:check`

Note that its important to run `generate` before `check`. These checks are performed in the CI pipeline.
Also note that some errors shown in `check` (non-silent mode) by `golicenses_check` are ignored (for more see [golicenses.sh.in](../../bazel/ci/golicenses.sh.in)).

### Query

* `bazel query //...` - list all targets
* `bazel query //subfolder` - list all targets in a subfolder
* `bazel cquery --output=files //subfolder:target` - get location of a build artifact

## Setup

### VS Code integration

You can continue to use the default Go language server and editor integration. This will show you different paths for external dependencies and not use the Bazel cache.
Alternatively, you can use [the go language server integration for Bazel](https://github.com/bazelbuild/rules_go/wiki/Editor-setup). This will use Bazel for dependency resolution and execute Bazel commands for building and testing.

### Command-line completion

[CLI completion for Bazel](https://bazel.build/install/completion) is available for Bash and zsh.

### Bash

When installing Bazel through the APT repository or Homebrew, completion scripts for bash should be installed automatically.

When building from source, you can install the completion script by adding the following line to your `~/.bashrc`:

```bash
source <path-to-constellation-repo>/bazel/bazel-complete.bash
```

### Zsh

When installing Bazel through the APT repository or Homebrew, completion scripts for zsh should be installed automatically. When using a heavily customized zsh config, you may need to follow [this workaround](https://bazel.build/install/completion).

When using Oh-My-Zsh, you can simply enable the [`zsh-autocomplete`](https://github.com/marlonrichert/zsh-autocomplete) plugin.

When building from source and not using Oh-My-Zsh, you can install the completion script as follows:

1. Locate the completion file, per default, it is located in `$HOME/.bazel/bin`
2. Add a file with the following to your `$fpath`

  ```zsh
  fpath[1,0]=~/.zsh/completion/
  mkdir -p ~/.zsh/completion/
  cp /path/from/above/step/_bazel ~/.zsh/completion
  ```

3. When installing for the first time, you may need to run `rm -f ~/.zcompdump; compinit` to rebuild the completion cache.
4. (Optional) Add the following to your `.zshrc`

  ```zsh
  # This way the completion script does not have to parse Bazel's options
  # repeatedly.  The directory in cache-path must be created manually.
  zstyle ':completion:*' use-cache on
  zstyle ':completion:*' cache-path ~/.zsh/cache
  ```
