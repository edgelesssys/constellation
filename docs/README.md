# Constellation Documentation

Published @ <https://docs.edgeless.systems/constellation> via `netlify`.

## Previewing

During edits you can preview your changes using the [`docusaurus`](https://docusaurus.io/docs/installation):

```sh
# requires node >=16.14
npm ci # Install pinned dependencies
npm run build
npm run serve
```

Browse to <http://localhost:3000/constellation> and choose the "Next" version in the top right.

## Release process

1. [Tagging a new version](https://docusaurus.io/docs/next/versioning#tagging-a-new-version)

    ```shell
    npm run docusaurus docs:version X.X
    ```

    When tagging a new version, the document versioning mechanism will:

    Copy the full `docs/` folder contents into a new `versioned_docs/version-[versionName]/` folder.
    Create a versioned sidebars file based from your current sidebar configuration (if it exists) - saved as `versioned_sidebars/version-[versionName]-sidebars.json`.
    Append the new version number to `versions.json`.
