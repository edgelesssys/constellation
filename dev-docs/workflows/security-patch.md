# Security Patch Workflow

This document describes how to patch vulnerabilities in Constellation.

## Guiding Principles

* Constellation vulnerabilities and security patches must be shared on need-to-know basis.
* A vulnerability is only fixed if a patch exists for all supported versions.
* Affected users must be informed about vulnerabilities affecting their setups.
* Vulnerabilities in Constellation should be fixed as quickly as possible.

## Vulnerability Report

Someone found a vulnerability in Constellation.
If they followed [SECURITY.md](/SECURITY.md), a Github Security Advisory might already exist.
Otherwise, now is the time to [create a draft](https://github.com/edgelesssys/constellation/security/advisories/new).
Make sure that the GHSA includes a problem statement that gives sufficient context and add domain experts as collaborators.

## Mitigation

Investigate possible mitigations for the vulnerability that don't require a patch release.
Such mitigations could include additional firewall settings, manual cluster configuration changes, etc.
Add all reasonable mitigation instructions to the GHSA.

If the vulnerability has already been disclosed publicly, the GHSA should also be disclosed at this stage.
Add an ETA for a patch release and proceed with [disclosure steps](#disclosing-the-ghsa).

## Fix

Sometimes a fix can be developed quickly for `main`, but correctly backporting it takes more time.
It may also happen that a proposed fix needs substantial work before merging.
In order to avoid premature disclosure of the vulnerability, while still allowing for collaboration, we use the GHSA's temporary repository.

1. On the drafted GHSA, create a temporary repository to work on a fix.
1. Develop a fix on a local branch, targeting `main`.
1. Manually run static checks (unit tests, linters, etc.).
1. Push the branch to the temporary fork and open a PR.
   This is necessary because the fork can't run Github Actions.
1. Solicit and incorporate feedback on the PR.
1. Manually test a fixed version, possibly including upgrade tests.

When the PR is ready, cherry-pick its commits to your local version of the release branch.
Repeat the steps above, but target the PR at the corresponding release branch.

Once PRs are ready for all necessary patches, hit the merge button on the GHSA.
This will merge all PRs, but the GHSA will stay in draft mode for now.

## Disclosing the GHSA

The following steps need to be performed by a repository admin.

1. Ensure that the GHSA text is in shape for publication.
   In particular, look out for any empty sections and placeholder text.
1. Fill in the `Affected Versions` and `Patched Versions` fields.
1. Check that the severity setting is accurate.
1. Credit external collaborators, e.g. by @-mention.
1. Hit the `Publish Advisory` button.
1. Tell Constellation users about the advisory (TODO: how?)
