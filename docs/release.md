# Release Checklist

This checklist will prepare `v1.3.0` from `v1.2.0`. Adjust your version numbers accordingly.

1. Merge ready PRs
2. Create a new branch `release/v1.3.0` to prepare the following things:
    1. Use [Build micro-service manual](https://github.com/edgelesssys/constellation/actions/workflows/build-micro-service-manual.yml) and run the pipeline once for each micro-service with the following parameters:
        * branch: `release/v1.3.0`
        * Container image tag: `v1.3.0`
        * Version of the image to build: `1.3.0`
        ```sh
        # Alternative from CLI
        gh workflow run build-micro-service-manual.yml --ref release/v1.3.0 -F microService=access-manager -F imageTag=v1.3.0 -F version=1.3.0
        gh workflow run build-micro-service-manual.yml --ref release/v1.3.0 -F microService=join-service -F imageTag=v1.3.0 -F version=1.3.0
        gh workflow run build-micro-service-manual.yml --ref release/v1.3.0 -F microService=kmsserver -F imageTag=v1.3.0 -F version=1.3.0
        gh workflow run build-micro-service-manual.yml --ref release/v1.3.0 -F microService=verification-service -F imageTag=v1.3.0 -F version=1.3.0
        ```
    2. Use [Build operator manual](https://github.com/edgelesssys/constellation/actions/workflows/build-operator-manual.yml) and run the pipeline once with the following parameters:
        * branch: `release/v1.3.0`
        * Container image tag: `v1.3.0`
        ```sh
        # Alternative from CLI
        gh workflow run build-operator-manual.yml --ref release/v1.3.0 -F imageTag=v1.3.0
        ```
    3. Review and update changelog with all changes since last release. [GitHub's diff view](https://github.com/edgelesssys/constellation/compare/v1.2.0...main) helps a lot!
    4. Update versions [versions.go](../internal/versions/versions.go#L33-L36) to `v1.3.0` and **push your changes**.
    5. Create a [production coreOS image](/.github/workflows/build-coreos.yml)
        ```sh
        gh workflow run build-coreos.yml --ref release/v0.0.1 -F debug=false -F coreOSConfigBranch=constellation
        ```
    6. Update [default images in config](/internal/config/images_enterprise.go)
    7. Merge this branch back to `main`
3. Run E2E to confirm stability and [generate measurements](/.github/workflows/e2e-test-manual.yml)
4. Create a new tag in `constellation` on `main`
    * `git tag v1.3.0`
    * Run [Release CLI](https://github.com/edgelesssys/constellation/actions/workflows/release-cli.yml) action on the tag
    * The previous step will create a draft release. Check build output for link to draft release. Review & approve.
5. Create a new tag in `constellation-docs`
    * `git tag v1.3.0`
