# dm-verity patch for CoreOS assembler

Constellation uses CoreOS as a base for OS images. While the images are mostly unmodified and can be built using the upstream CoreOS assembler, small modifications to the assembler are required to support dm-verity for the root filesystem.

Checkout the CoreOS assembler source code [from the upstream repo](https://github.com/coreos/coreos-assembler) using the commit ID specified in the [Makefile](Makefile)

```shell-session
make clone
```

Apply the patch:

```shell-session
make patch
```

Now you can make changes to the coreos-assembler and compile it using the included `Dockerfile`:

```shell-session
make containerimage
```

Once you are done, create a new patch file (within `3rdparty/coreos-assembler/build/coreos-assembler`):
```shell-session
git diff HEAD^ > ../../verity.patch
```

## Building the CoreOS assembler container image

```shell-session
make
```

The resulting container image will be tagged as `localhost/coreos-assembler`.
