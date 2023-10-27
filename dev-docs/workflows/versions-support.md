# Versions support

This explains the support and compatibility of the different versions that are part of the config.

## Microservice version

(the version that is referenced in the config during `constellation apply`) must match the CLI version exactly (down to the version suffix). When planning an upgrade, the source version (current version before the upgrade is applied) may be up to one minor version older (allow upgrades from v2.8.x to v2.9.x).
The reason is this: The CLI embeds exactly one version of the helm charts (the version of the CLI). So whenever we deploy the helm charts, we can only use the embedded chart version.
The source version can only be older by one minor version to ensure we can perform safe upgrades where we can reasonably test what microservice versions will interact with other versions of CLI, OS images, other microservices.

## Kubernetes version

(the version that is referenced in the config during `constellation apply`) must match one of the embedded k8s versions. We have a hardcoded list that is embedded in the CLI.
When planning an upgrade, the source k8s version may be up to one minor version older (allow upgrades from k8s 1.25.x to 1.26.x).
The reason is this: The CLI embeds components for kubernetes version for the hardcoded list of supported k8s versions. So whenever we deploy a k8s version (during `constellation apply`), the k8s components for the target version need to be available.
The source version drift is a general k8s recommendation: k8s was designed to withstand a version drift of one minor version in each component during a rolling upgrade.

## Image version

(the version that is referenced in the config during `constellation apply`) must have the same major and minor version (but can have a different patch or suffix) when upgrading compared to the cli. Examples: CLI v2.8.2 could correctly upgrade to OS image v2.8.1 or v2.8.2-nightly-foo-bar.
When planning an upgrade, the source image version may be up to one minor version older (allow upgrades from v2.8.x to 2.9.x).
The reason is this: The CLI uses an API endpoint to query configuration and measurements for a configured target image. This generally allows for arbitrary image versions to be referenced/deployed by the CLI. However, to ensure safe upgrades, we always upgrade by at most one minor version and ensure that the CLI is at the same minor version that we target.
This allows us to test upgrades with only a reasonable amount of moving parts.
Patch version upgrades of OS images are not allowed to have any breaking changes. They are designed to deliver bug and security fixes only and should otherwise be 1:1 compatible.
