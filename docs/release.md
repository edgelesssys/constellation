# Release Checklist

This checklist will prepare `v1.3.0` from `v1.2.0`. Adjust your version numbers accordingly.

1. Merge ready PRs
2. Create a new branch `release/v1.3.0` to prepare the following things:
    1. Use [Build micro-service manual](https://github.com/edgelesssys/constellation/actions/workflows/build-micro-service-manual.yml) and run the pipeline once for each micro-service with the following parameters:
        * branch: `release/v1.3.0`
        * Container image tag: `v1.3.0`
        * Version of the image to build: `1.3.0`
    2. Review and update changelog with all changes since last release. [GitHub's diff view](https://github.com/edgelesssys/constellation/compare/v1.2.0...main) helps a lot!
    3. Update versions [versions.go](../internal/versions/versions.go#L33-L36) to `v1.3.0`
    4. Create a [production coreOS image](/.github/workflows/build-coreos.yml)
    5. Update [default images in config](/internal/config/config.go)
    6. Merge this branch back to `main`
3. Run E2E to confirm stability and [generate measurements](/.github/workflows/e2e-test-manual.yml)
4. Create a new tag in `constellation` on `main`
    * `git tag v1.3.0`
    * Run [Build CLI](https://github.com/edgelesssys/constellation/actions/workflows/build-cli.yml) action on the tag
    * The previous step will create a draft release. Check build output for link to draft release. Review & approve.
5. Create a new tag in `constellation-docs`
    * `git tag v1.3.0`
