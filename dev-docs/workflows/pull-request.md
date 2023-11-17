# Pull Request (PR)

## Process

Submissions should remain focused in scope and avoid containing unrelated commits.
For pull requests, we employ the following workflow:

1. Fork the repository to your own GitHub account
2. Create a branch locally with a descriptive name
3. Commit changes to the branch
4. Write your code according to our [development guidelines](../conventions.md)
5. Push changes to your fork
6. Clean up your commit history
7. Open a PR in our repository and summarize the changes in the description

### Major changes and feature requests

You should discuss larger changes and feature requests with the maintainers. Please open an issue describing your plans.

[Run CI e2e tests](github-actions.md)

## Conventions

### Title

Our changelog is generated from PR titles, so please stick to the naming convention.

The PR title should be structured in one of the following ways:

```
<module>: <title>
```

Where the `<module>` is

* the top level directory of the microservice or component, e.g., `joinservice`, `disk-mapper`, `upgrade-agent` but also `docs` and `rfc`
* in internal, the second level directory
* `deps` for dependency upgrades
* `ci` for things that are CI related

and `<title>` is all lower case (except proper names, including acronyms).
Ticket numbers shouldn’t be part of the title.

In case the scope of your PR is too wide, use the alternative format.

```
<Title>
```

and `<Title>` starts with a capital letter.

### Labels

The labels are used for changelog generation (targeted at constellation users), so select the label with this purpose in mind.
To exclude the PR from changelog only use these labels:

* `no changelog` / `dependencies`

 The changelog generation is described [here](https://docs.github.com/en/repositories/releasing-projects-on-github/automatically-generated-release-notes). This is our configuration [release.yml](/.github/release.yml).
