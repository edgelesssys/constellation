# Release Checklist

This checklist will prepare `v1.3.0` from `v1.2.0`. Adjust your version numbers accordingly.

1. Merge ready PRs
2. Create docs release (new major or minor release)

    ```sh
    cd docs
    npm install
    npm run docusaurus docs:version 1.3
    # push upstream via PR
    ```

3. Create a new branch `release/v1.3` (new minor version) or use the existing one (new patch version)
4. On this branch, prepare the following things:
    1. (new patch version) `cherry-pick` (only) the required commits from `main`
    2. Use [Build micro-service manual](https://github.com/edgelesssys/constellation/actions/workflows/build-micro-service-manual.yml) and run the pipeline once for each micro-service with the following parameters:
        * branch: `release/v1.3`
        * Container image tag: `v1.3.0`
        * Version of the image to build: `1.3.0`

       ```sh
       ver=1.3.0
       ```

        ```sh
        minor=$(echo $ver | cut -d '.' -f 1,2)
        gcpVer=$(echo $ver | tr "." "-")
        echo $minor # should be 1.3
        echo $gcpVer # should be 1-3-0
        ```

        ```sh
        gh workflow run build-micro-service-manual.yml --ref release/v$minor -F microService=access-manager -F imageTag=v$ver -F version=$ver --repo edgelesssys/constellation
        gh workflow run build-micro-service-manual.yml --ref release/v$minor -F microService=join-service -F imageTag=v$ver -F version=$ver --repo edgelesssys/constellation
        gh workflow run build-micro-service-manual.yml --ref release/v$minor -F microService=kmsserver -F imageTag=v$ver -F version=$ver --repo edgelesssys/constellation
        gh workflow run build-micro-service-manual.yml --ref release/v$minor -F microService=verification-service -F imageTag=v$ver -F version=$ver --repo edgelesssys/constellation
        ```

    3. Use [Build operator manual](https://github.com/edgelesssys/constellation/actions/workflows/build-operator-manual.yml) and run the pipeline once with the following parameters:
        * branch: `release/v1.3`
        * Container image tag: `v1.3.0`

        ```sh
        # Alternative from CLI
        gh workflow run build-operator-manual.yml --ref release/v$minor -F imageTag=v$ver --repo edgelesssys/constellation
        ```

    4. Review and update changelog with all changes since last release. [GitHub's diff view](https://github.com/edgelesssys/constellation/compare/v2.0.0...main) helps a lot!
       1. Rename the "Unreleased" heading to "[v1.3.0] - YYYY-MM-DD" and link the version to the upcoming release tag.
       2. Create a new block for unreleased changes
    5. Update project version in [CMakeLists.txt](/CMakeLists.txt) to `1.3.0` (without v).
    6. When the microservice builds are finished update versions in [versions.go](../../internal/versions/versions.go#L33-L39) to `v1.3.0`, **add the container hashes** and **push your changes**.
    7. Create a [production OS image](/.github/workflows/build-coreos.yml)

        ```sh
        gh workflow run build-os-image.yml --ref release/v$minor -F debug=false -F imageVersion=v$ver
        ```

    8. Update [default images in config](/internal/config/images_enterprise.go)
    9. Run manual E2E tests using [Linux](/.github/workflows/e2e-test-manual.yml) and [macOS](/.github/workflows/e2e-test-manual-macos.yml) to confirm functionality and stability.

        ```sh
        sono='--plugin e2e --plugin-env e2e.E2E_FOCUS="\[Conformance\]" --plugin-env e2e.E2E_SKIP="for service with type clusterIP|HostPort validates that there is no conflict between pods with same hostPort but different hostIP and protocol" --plugin https://raw.githubusercontent.com/vmware-tanzu/sonobuoy-plugins/master/cis-benchmarks/kube-bench-plugin.yaml --plugin https://raw.githubusercontent.com/vmware-tanzu/sonobuoy-plugins/master/cis-benchmarks/kube-bench-master-plugin.yaml'
        gh workflow run e2e-test-manual.yml --ref release/v$minor -F cloudProvider=azure -F machineType=Standard_DC4as_v5 -F sonobuoyTestSuiteCmd="$sono" -F osImage=/CommunityGalleries/ConstellationCVM-b3782fa0-0df7-4f2f-963e-fc7fc42663df/Images/constellation/Versions/$ver -F isDebugImage=false
        gh workflow run e2e-test-manual-macos.yml --ref release/v$minor -F cloudProvider=azure -F machineType=Standard_DC4as_v5 -F sonobuoyTestSuiteCmd="$sono" -F osImage=/CommunityGalleries/ConstellationCVM-b3782fa0-0df7-4f2f-963e-fc7fc42663df/Images/constellation/Versions/$ver -F isDebugImage=false
        gh workflow run e2e-test-manual.yml --ref release/v$minor -F cloudProvider=gcp -F machineType=n2d-standard-4 -F sonobuoyTestSuiteCmd="$sono" -F osImage=projects/constellation-images/global/images/constellation-v$gcpVer -F isDebugImage=false
        gh workflow run e2e-test-manual-macos.yml --ref release/v$minor -F cloudProvider=gcp -F machineType=n2d-standard-4 -F sonobuoyTestSuiteCmd="$sono" -F osImage=projects/constellation-images/global/images/constellation-v$gcpVer -F isDebugImage=false
        ```

    10. [Generate measurements](/.github/workflows/generate-measurements.yml) for the images on each CSP.

        ```sh
           gh workflow run generate-measurements.yml --ref release/v$minor -F cloudProvider=azure -F osImage=/CommunityGalleries/ConstellationCVM-b3782fa0-0df7-4f2f-963e-fc7fc42663df/Images/constellation/Versions/$ver -F isDebugImage=false
           gh workflow run generate-measurements.yml --ref release/v$minor -F cloudProvider=gcp -F osImage=projects/constellation-images/global/images/constellation-v$gcpVer -F isDebugImage=false
        ```

    11. Create a new tag on this release branch
        ```sh
        git tag v$ver
        git tags --push
        ```

        * Run [Release CLI](https://github.com/edgelesssys/constellation/actions/workflows/release-cli.yml) action on the tag

        ```sh
        gh workflow run release-cli.yml --ref v$ver
        ```

        * The previous step will create a draft release. Check build output for link to draft release. Review & approve.
5. Follow [export flow (INTERNAL)](https://github.com/edgelesssys/wiki/blob/master/documentation/constellation/customer-onboarding.md#manual-export-and-import) to make image available in S3 for trusted launch users.
6. To bring updated version numbers and other changes (if any) to main, create a new branch `feat/release` from `release/v1.3`, rebase it onto main, and create a PR to main
7. Milestones management
   1. Create a new milestone for the next release
   2. Add the next release manager and an approximate release date to the milestone description
   3. Close the milestone for the release
   4. Move open issues and PRs from closed milestone to next milestone
8. If the release is a minor version release, create an empty commit on main and tag it as the start of the next pre-release phase.
    ```sh
    nextMinorVer=$(echo $ver | awk -F. -v OFS=. '{$2 += 1 ; print}')
    git checkout main
    git pull
    git commit --allow-empty -m "Start v$nextMinorVer-pre"
    git tag v$nextMinorVer-pre
    git push origin main v$nextMinorVer-pre
    ```
