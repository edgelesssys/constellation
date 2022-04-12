## Proto generation

To generate Go source files from proto, we use docker.

The following command will generate Go source code files in docker and save the output to the relevant directory.
Run this once every time you make any changes or additions to the `.proto` files.
Add the generated `.go` files, and any changes to the `.proto` files, to your branch before creating a PR.

```bash
DOCKER_BUILDKIT=1 docker build -o .. -f Dockerfile.gen-proto ..
```
