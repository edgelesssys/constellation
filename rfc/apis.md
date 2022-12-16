# Constellation APIs (v1)

## Base

The API base starts with `constellation` followed by the API version.
At this moment, the only valid API version is `v1`:

```
/constellation/v1/
```

## API groups

The API version is followed by the API group. Possible values are:

- [`versions`: version information for Constellation components](version-api.md)
- [`image`: metadata for individual Constellation OS images](image-api.md)
  - `info`: image reference lookup for each cloud provider and additional metadata
  - `measurements`: TPM measurements for Constellation OS images
  - `raw`: raw OS images
  - `sbom`: SBOMs for Constellation OS images

There may be more API groups in the future (e.g. `cli`)

## API paths overview

- [`/constellation/v1/ref/<ref>/stream/<stream>/versions/latest/<kind>.json`](version-api.md#latest)
- [`/constellation/v1/ref/<ref>/stream/<stream>/versions/major/<base>/<kind>.json`](version-api.md#major-to-minor-version-list)
- [`/constellation/v1/ref/<ref>/stream/<stream>/versions/minor/<base>/<kind>.json`](version-api.md#minor-to-patch-version-list)
- [`/constellation/v1/ref/<ref>/stream/<stream>/image/<version>/info.json`](image-api.md#image-lookup-table)
- [`/constellation/v1/ref/<ref>/stream/<stream>/image/<version>/sbom.<format>.json`](image-api.md)
- [`/constellation/v1/ref/<ref>/stream/<stream>/image/<version>/csp/<csp>/measurements.json`](image-api.md)
- [`/constellation/v1/ref/<ref>/stream/<stream>/image/<version>/csp/<csp>/measurements.json.sig`](image-api.md)
- [`/constellation/v1/ref/<ref>/stream/<stream>/image/<version>/csp/<csp>/measurements.image.json`](image-api.md)
- [`/constellation/v1/ref/<ref>/stream/<stream>/image/<version>/csp/<csp>/image.raw`](image-api.md)

## API path identifiers  `ref`, `stream` and `version`

### Meaning of `ref`, `stream` and `version` for resource names

Components in this API are identified by `ref`, `stream` and `version`.

- a `ref` that is either a normalized git branch name or the special value `-`. Defaults to `-`.
- a `stream` that determines the specific type of the resource. Defaults to `stable`.
- a `version` that is a semantic version.

The special `ref` value `-` (dash) is reserved for releases. Every other value is a branch name where special characters are replaced by dashes.

Streams are used to distinguish different types of resources. Depending on the kind of resource and `ref`, only a subset of values might be allowed.

The `version` is always a valid semantic version. For the special stream `stable` it is guaranteed to be a semantic version `v<major>.<minor>.<patch>`,
for other stream values, it is a semantic version that may contain a suffix after the patch: `v<major>.<minor>.<patch>-<suffix>`.

The use of semantic versions and pseudo versions tries to match the versioning used in Go. If in doubt,
consult the module reference of Go, especially the sections about [versions](https://go.dev/ref/mod#versions)
and [pseudo-versions](https://go.dev/ref/mod#pseudo-versions).

The [pesudo-version tool](../hack/pseudo-version) can generate a valid pseudo-version for your current head.

### Consisten API path prefix of `ref` and `stream`

For API calls, paths will always start with `ref` and `stream`:

```
<base>/ref/<ref>/stream/<stream>/
```

### Meaning of `ref` and `stream` for images

The special value `-` (dash) for images is used for releases. They are always built from the corresponding release branch (and tag).
Every other value refers to a normalized branch name.

For the `ref` `-` (dash), the `stream` value can be one of the following constants:

- `stable`: Built with default settings (non-debug), for end users and with all security features enabled
- `console`: Built with default settings (non-debug), allows access to the serial console
- `debug`: Image containing the debugd, allows access to the serial console

For other `ref` values, the `stream` value can be one of the following constants:

- `nightly`: Built with default settings (non-debug), with all security features enabled
- `console`: Built with default settings (non-debug), allows access to the serial console
- `debug`: Image containing the debugd, allows access to the serial console

## Short paths for resource identification

Short paths allow an easier version handling, e.g., in end user configuration files or CLI tools.
When using the default values, short paths omit either the `ref` (if set to `-`) or the `ref` and `stream`
parts (if set to `-` and `stable`):

- `ref/-/stream/stable/<version>` is equivalent to the short path `<version>`
- `ref/-/stream/<stream>/<version>` is equivalent to the short path `stream/<stream>/<version>`

Resource group can be identified by `ref/<ref>/stream/<stream>/<version>` and allow for short form encoding as explained above.
Resources using this encoding include:

- `version`: version information for Constellation components
- `image`: metadata for individual Constellation OS images

## Examples

Release v2.3.0 would use the following image name for the default, end-user image:

```
 short form:                     v2.3.0
  long form: ref/-/stream/stable/v2.3.0
```

A debug image created from the same release:

```
 short form:       stream/debug/v2.3.0
  long form: ref/-/stream/debug/v2.3.0
```

A debug image on the branch `feat/xyz`:

```
ref/feat-xyz/stream/debug/v2.4.0-pre.0.20230922011244-0744d001aa84
```

A nightly image on `main`:

```
ref/main/stream/nightly/v2.4.0-pre.0.20230922011244-0744d001aa84
```

A debug image on `main`:

```
ref/main/stream/debug/v2.4.0-pre.0.20230922011244-0744d001aa84
```
