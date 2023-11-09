# Bump Go version
`govulncheck` from the bazel `check` target will fail if the Go version is outdated.

## Steps

1. Replace "1.xx.x" with the new version (see [example](https://github.com/edgelesssys/constellation/commit/9e1a0c06bfda0171958f0776633a9a53f521144d))
2. Update the nix hash

    Once updated run `bazel run //:tidy` and you will see a failure such as:

    ```
        > error: hash mismatch in fixed-output derivation '/nix/store/r85bdj6vrim7m5vlybdmzgca7d0kcb4n-go1.21.4.src.tar.gz.drv':
      >          specified: sha256-GG8rb4yLcE5paCGwmrIEGlwe4T3LwxVqE63PdZMe5Ig=
      >             got:    sha256-R7Jqg9K2WjwcG8rOJztpvuSaentRaKdgTe09JqN714c=
    ```
    Simple replace the hash with the got value.
