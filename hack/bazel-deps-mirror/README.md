# Bazel deps mirror

This directory contains tooling to automatically mirror the dependencies of a Bazel project into the Constellation CDN at `https://cdn.confidential.cloud/`.

The tool searches for various rules in the WORKSPACE.bzlmod file and all loaded .bzl files.
It has the following commands:

- check: checks if the dependencies all have a mirror URL and optionally checks if the mirror really returns the expected file
- mirror: mirrors all dependencies that don't have a mirror URL yet. Also normalizes the `urls` attribute of rules
