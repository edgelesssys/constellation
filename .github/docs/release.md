# Release Checklist

This checklist will prepare `v1.3.0` from `v1.2.0`. Adjust your version numbers accordingly.

1. Merge ready PRs
2. Create docs release (new major or minor release)
    ```sh
    cd docs
    npm run docusaurus docs:version 1.3
    # push upstream via PR
    ```
3. On the [CoreOS config repo](https://github.com/edgelesssys/constellation-fedora-coreos-config), create two new branches `release/v1.3`, `stream/v1.3` (new minor version) or use the existing ones (new patch version).
   The release branch contains the squashed changeset and is branched from main while the stream branch contains the rebased changesets on top of the latest upstream changes.
   [Consult this guide on rebasing forks (INTERNAL)](https://github.com/edgelesssys/wiki/blob/master/documentation/rebasing_forks.md#managing-release-branches) for guidance.
4. Create a new branch `release/v1.3` (new minor version) or use the existing one (new patch version)
5. On this branch, prepare the following things:
    1. (new patch version) `cherry-pick` (only) the required commits from `main`
    2. Use [Build micro-service manual](https://github.com/edgelesssys/constellation/actions/workflows/build-micro-service-manual.yml) and run the pipeline once for each micro-service with the following parameters:
        * branch: `release/v1.3`
        * Container image tag: `v1.3.0`
        * Version of the image to build: `1.3.0`
        ```sh
        # Alternative from CLI
        gh workflow run build-micro-service-manual.yml --ref release/v1.3 -F microService=access-manager -F imageTag=v1.3.0 -F version=1.3.0
        gh workflow run build-micro-service-manual.yml --ref release/v1.3 -F microService=join-service -F imageTag=v1.3.0 -F version=1.3.0
        gh workflow run build-micro-service-manual.yml --ref release/v1.3 -F microService=kmsserver -F imageTag=v1.3.0 -F version=1.3.0
        gh workflow run build-micro-service-manual.yml --ref release/v1.3 -F microService=verification-service -F imageTag=v1.3.0 -F version=1.3.0
        ```
    3. Use [Build operator manual](https://github.com/edgelesssys/constellation/actions/workflows/build-operator-manual.yml) and run the pipeline once with the following parameters:
        * branch: `release/v1.3`
        * Container image tag: `v1.3.0`
        ```sh
        # Alternative from CLI
        gh workflow run build-operator-manual.yml --ref release/v1.3 -F imageTag=v1.3.0
        ```
    4. Review and update changelog with all changes since last release. [GitHub's diff view](https://github.com/edgelesssys/constellation/compare/v1.2.0...main) helps a lot!
    5. Update project version in [CMakeLists.txt](/CMakeLists.txt) to `1.3.0` (without v).
    6. Update versions [versions.go](../../internal/versions/versions.go#L33-L39) to `v1.3.0` and **push your changes**.
    7. Create a [production coreOS image](/.github/workflows/build-coreos.yml)
        ```sh
        gh workflow run build-coreos.yml --ref release/v1.3 -F debug=false -F coreOSConfigBranch=release/v1.3
        ```
    8. Update [default images in config](/internal/config/images_enterprise.go)
    9. Run E2E tests to confirm stability and [generate measurements](/.github/workflows/e2e-test-manual.yml)
        ```sh
        gh workflow run e2e-test-manual.yml --ref release/v1.3 -F workerNodesCount=2 -F controlNodesCount=1 -F autoscale=false -F cloudProvider=azure -F machineType=Standard_DC4as_v5 -F sonobuoyTestSuiteCmd="--mode quick" -F kubernetesVersion=1.23 -F coreosImage=/CommunityGalleries/ConstellationCVM-b3782fa0-0df7-4f2f-963e-fc7fc42663df/Images/constellation/Versions/1.3.0 -F isDebugImage=false
        gh workflow run e2e-test-manual.yml --ref release/v1.3 -F workerNodesCount=2 -F controlNodesCount=1 -F autoscale=false -F cloudProvider=gcp -F machineType=n2d-standard-4 -F sonobuoyTestSuiteCmd="--mode quick" -F kubernetesVersion=1.23 -F coreosImage=projects/constellation-images/global/images/constellation-v1-3-0 -F isDebugImage=false
        ```
    10. Create a new tag on this release branch
        * `git tag v1.3.0`
        * Run [Release CLI](https://github.com/edgelesssys/constellation/actions/workflows/release-cli.yml) action on the tag
        ```sh
        gh workflow run release-cli.yml --ref v1.3.0
        ```
        * The previous step will create a draft release. Check build output for link to draft release. Review & approve.
6. Folow [export flow](https://github.com/edgelesssys/wiki/blob/master/documentation/constellation/customer-onboarding.md#manual-export-and-import) to make image available in S3 for trusted lanuch users.
7. To bring updated version numbers and other changes (if any) to main, create a new branch `feat/release` from `release/v1.3`, rebase it onto main, and create a PR to main
