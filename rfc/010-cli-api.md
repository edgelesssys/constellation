# RFC 010: CLI compatibility information

The CLI API provides information about the compatibility of the Constellation CLI and other components of the Constellation ecosystem such as Kubernetes versions.

## CLI API Endpoints

The build pipeline produces artifacts for compatibility information that are uploaded to S3 and can be accessed via HTTP.
The artifacts are organized in a directory structure that allows to look up the compatibility for a given Constellation version.

The following HTTP endpoints are available:

- `GET /constellation/v1/ref/<REF>/stream/<STREAM>/<VERSION>/cli/`
  - [`info.json` returns the CLI compatibility information artifact for the given Constellation version.](#cli-lookup-table)

## CLI information artifact

The CLI compatibility information artifact is a JSON file that maps the image name consisting of `ref`, `stream` and `version` to the corresponding CLI version and it's compatibility information:

```
/constellation/v1/ref/<REF>/stream/<STREAM>/<VERSION>/cli/info.json
```

```json
{
  "version": "<VERSION>",
  "ref": "<REF>",
  "stream": "<STREAM>",
  "kubernetes": ["v1.1.23", "v1.1.24", "v1.1.25"]
}
```

This shows that the Constellation CLI version `<VERSION>` is compatible with Kubernetes versions `v1.1.23`, `v1.1.24` and `v1.1.25`.
