# Release Checklist

This checklist will prepare `v1.3.0` from `v1.2.0`. Adjust your version numbers accordingly.

1. Merge ready PRs
2. Use [Build micro-service manual](https://github.com/edgelesssys/constellation/actions/workflows/build-micro-service-manual.yml) and run the pipeline once for each micro-service with the following parameters:
    * branch: `main`
    * Container image tag: `v1.3.0`
    * Version of the image to build: `1.3.0`
3. Create a new branch to prepare the following things:
    1. Review and update changelog with all changes since last release. [GitHub's diff view](https://github.com/edgelesssys/constellation/compare/v1.2.0...main) helps a lot!
    2. Update versions [images.go](../coordinator/kubernetes/k8sapi/resources/images.go) to `v1.3`. Omit patch version so containers pick up patch level updates automatically.
    3. Merge this branch
4. Create a new tag in `constellation`
    * `git tag v.1.3.0`
    * Run [Build CLI](https://github.com/edgelesssys/constellation/actions/workflows/build-cli.yml) action on the tag
    * Check build output for link to draft release. Review & approve.
5. Create a new tag in `constellation-docs`
    * `git tag v.1.3.0`
