# Local RPM repo

This folder contains helper scripts to create a local RPM repository for the image build process.
By default, it will download pinned versions of the packages from a container image.

The repository is created in the `repo` subfolder.

## Usage

```sh
make pull  # Download pinned packages from container image
make repo  # Create local RPM repository
```

## Updating packages

```sh
make update                 # Recreate the SHA256SUMS file with upstream packages
make repo                   # Create local RPM repository
make testrepo               # Test the repository by creating an OS image using only the local repository
make push                   # Push new packages to the container image
git add TAG repo/SHA256SUMS # Commit the newly pinned packages
```
