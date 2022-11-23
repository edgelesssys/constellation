# Update API

The update API should expose easy, straightforward, extensible and forward compatible version information to the Constellation CLI (and possibly more consumers).

Design goals:

- Simple
- Can be implemented using static HTTP file server
- Generic over different kinds of resources
    - OS image versions
    - Microservice versions
    - CLI versions
    - Kubernetes versions

The folowing HTTP endpoints are available:

- `GET /constellation/v1/updates/latest/` contains files showing the latest version available
    - `image.json` contains the latest image version
    - `microservice.json` contains the latest microservice version
    - `cli.json` contains the latest cli version
    - `kubernetes.json` contains the latest supported version of Kubernetes
- `GET /constellation/v1/updates/major/<major-version>/` contains files with version information for this major version
    - `image.json` contains a list of all minor image versions that belong to a major version
    - `microservice.json` contains a list of all minor microservice versions that belong to a major version
    - `cli.json` contains a list of all minor cli versions that belong to a major version
    - `kubernetes.json` contains a list of all supported minor version of Kubernetes that belong to a major version
- `GET /constellation/v1/updates/minor/<minor-version>/` contains files with version information for this minor version
    - `image.json` contains a list of all patch image versions that belong to a minor version
    - `microservice.json` contains a list of all patch microservice versions that belong to a minor version
    - `cli.json` contains a list of all patch cli versions that belong to a minor version
    - `kubernetes.json` contains a list of all supported patch version of Kubernetes that belong to a minor version

## Examples

This shows the version information for the latest OS image. The file could be extended to include more metadata.

```
https://cdn.confidential.cloud/constellation/v1/updates/latest/image.json
```

```json
{
    "version": "v2.3.0"
}
```

This shows the version information for the latest Kubernetes release. The file could be extended to include more metadata.

```
https://cdn.confidential.cloud/constellation/v1/updates/latest/kubernetes.json
```

```json
{
    "version": "v1.25.4"
}
```

This shows a list of all minor releases of the microservices for the major version `v2`. The file could be extended to include more metadata.

```
https://cdn.confidential.cloud/constellation/v1/updates/major/v2/microservice.json
```

```json
{
    "versions": ["v2.0", "v2.1", "v2.2", "v2.3"]
}
```

This shows a list of all patch releases of the CLI for the minor version `v2.3`. The file could be extended to include more metadata.

```
https://cdn.confidential.cloud/constellation/v1/updates/minor/v2.3/cli.json
```

```json
{
    "versions": ["v2.3.0", "v2.3.1", "v2.3.2", "v2.3.3"]
}
```

## Update discovery

The update API can be used to find new versions efficiently and can be used to perform simple selection of available upgrades using semantic versioning.
Below are some example use cases.

### Inform about updates

The CLI can query the `latest` endpoint to inform users about new releases of the CLI itself.

### Propose compatible image updates

The CLI can query the `minor` endpoint to retrieve a list of all patch releases of a Constellation OS image. This can be used to only select compatible versions for image updates if combined with a downgrade protection (only allow updating to image versions that are newer than the one currently in use).


## Possible future extensions

- Explicit compatbility information between different resource kinds
    - Example: OS image `v2.3.4` requires microservice version > `v2.3.1`
    - Allows fine grained control in case of unexpected incompatibility
- Version metadata
    - Changelogs
    - Deprecation warnings
    - Security information