# Supply chain levels for software artifacts (SLSA) adoption

[Supply chain Levels for Software Artifacts, or SLSA (salsa)](https://slsa.dev/) is a framework for improving and grading a project's build system and engineering processes. SLSA focuses on security improvements for source code storage as well as build system definition, execution, and observation. SLSA is structured in [four levels](https://slsa.dev/spec/v0.1/levels). This page describes the adoption of SLSA for Constellation.

:::info
SLSA is still in alpha status. The presented levels and their requirements might change in the future. We will adopt any changes into our engineering processes, as they get defined.
:::

## Level 1 - Adopted

**[Build - Scripted](https://slsa.dev/spec/v0.1/requirements#scripted-build)**

All build steps are automated via [Bazel](https://github.com/edgelesssys/constellation/tree/main/bazel/ci) and [GitHub Actions](https://github.com/edgelesssys/constellation/tree/main/.github).

**[Provenance - Available](https://slsa.dev/spec/v0.1/requirements#available)**

Provenance for the CLI is generated using the [slsa-github-generator](https://github.com/slsa-framework/slsa-github-generator).

## Level 2 - Adopted

**[Source - Version Controlled](https://slsa.dev/spec/v0.1/requirements#version-controlled)**

Constellation is hosted on GitHub using git.

**[Build - Build Service](https://slsa.dev/spec/v0.1/requirements#build-service)**

All builds are carried out by [GitHub Actions](https://github.com/edgelesssys/constellation/tree/main/.github).

**[Provenance - Authenticated](https://slsa.dev/spec/v0.1/requirements#authenticated)**

Provenance for the CLI is signed using the [slsa-github-generator](https://github.com/slsa-framework/slsa-github-generator). Learn [how to verify the CLI](../workflows/verify-cli.md) using the signed provenance, before using it for the first time.

**[Provenance - Service Generated](https://slsa.dev/spec/v0.1/requirements#service-generated)**

Provenance for the CLI is generated using the [slsa-github-generator](https://github.com/slsa-framework/slsa-github-generator) in GitHub Actions.

## Level 3 - Adopted

**[Source - Verified History](https://slsa.dev/spec/v0.1/requirements#verified-history)**

The [Edgeless Systems](https://github.com/edgelesssys) GitHub organization [requires two-factor authentication](https://docs.github.com/en/organizations/keeping-your-organization-secure/managing-two-factor-authentication-for-your-organization/requiring-two-factor-authentication-in-your-organization) for all members.

**[Source - Retained Indefinitely](https://slsa.dev/spec/v0.1/requirements#retained-indefinitely)**

Since we use GitHub to host the repository, an external person can't modify or delete the history. Before a pull request can be merged, an explicit approval from an [Edgeless Systems](https://github.com/edgelesssys) team member is required.

The same holds true for changes proposed by team members. Each change to `main` needs to be proposed via a pull request and requires at least one approval.

The [Edgeless Systems](https://github.com/edgelesssys) GitHub organization admins control these settings and are able to make changes to the repository's history should legal requirements necessitate it. These changes require two-party approval following the obliterate policy.

**[Build - Build as Code](https://slsa.dev/spec/v0.1/requirements#build-as-code)**

All build files for Constellation are stored in [the same repository](https://github.com/edgelesssys/constellation/tree/main/.github).

**[Build - Ephemeral Environment](https://slsa.dev/spec/v0.1/requirements#ephemeral-environment)**

All GitHub Action workflows are executed on [GitHub-hosted runners](https://docs.github.com/en/actions/using-github-hosted-runners/about-github-hosted-runners). These runners are only available during workflow.

We currently don't use [self-hosted runners](https://docs.github.com/en/actions/hosting-your-own-runners/about-self-hosted-runners).

**[Build - Isolated](https://slsa.dev/spec/v0.1/requirements#isolated)**

As outlined in the previous section, we use GitHub-hosted runners, which provide a new, isolated and ephemeral environment for each build.

Additionally, the [SLSA GitHub generator](https://github.com/slsa-framework/slsa-github-generator#generation-of-provenance) itself is run in an isolated workflow with the artifact hash as defined inputs.

**[Provenance - Non-falsifiable](https://slsa.dev/spec/v0.1/requirements#non-falsifiable)**

As outlined by [SLSA GitHub generator](https://github.com/slsa-framework/slsa-github-generator) it already fulfills the non-falsifiable requirements for SLSA Level 3. The generated provenance is signed using [sigstore](https://sigstore.dev/) with an OIDC based proof of identity.

## Level 4 - In Progress

We strive to adopt certain aspect of SLSA Level 4 that support our engineering process. At the same time, SLSA is still in alpha status and the biggest changes to SLSA are expected to be around Level 4.
