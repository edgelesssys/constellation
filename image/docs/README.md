# Fedora CoreOS Assembler

We use the [Fedora CoreOS Assembler](https://coreos.github.io/coreos-assembler/) to build the base image for Constellation nodes.

## Setup

Prerequisites: `podman` and `qemu-kvm` are installed, nested virtualization is enabled.
Make sure your user is allowed read and write access on `/dev/kvm`.
If the device is not mounted in the container try the following command, and restart the container:
``` shell
sudo chmod 666 /dev/kvm
```

* Pull the assembler container image

    ``` shell
    podman pull quay.io/coreos-assembler/coreos-assembler
    ```

* Create a working directory on your host system

    ``` shell
    mkdir fcos && cd fcos
    ```

* Set up a bash alias

    Add the following to your `.bashrc` to easily start the image assembler using `cosa`:
    ``` bash
    cosa() {
        env | grep COREOS_ASSEMBLER
        local -r COREOS_ASSEMBLER_CONTAINER_LATEST="quay.io/coreos-assembler/coreos-assembler:latest"
        if [[ -z ${COREOS_ASSEMBLER_CONTAINER} ]] && $(podman image exists ${COREOS_ASSEMBLER_CONTAINER_LATEST}); then
            local -r cosa_build_date_str="$(podman inspect -f "{{.Created}}" ${COREOS_ASSEMBLER_CONTAINER_LATEST} | awk '{print $1}')"
            local -r cosa_build_date="$(date -d ${cosa_build_date_str} +%s)"
            if [[ $(date +%s) -ge $((cosa_build_date + 60*60*24*7)) ]] ; then
                echo -e "\e[0;33m----" >&2
                echo "The COSA container image is more that a week old and likely outdated." >&2
                echo "You should pull the latest version with:" >&2
                echo "podman pull ${COREOS_ASSEMBLER_CONTAINER_LATEST}" >&2
                echo -e "----\e[0m" >&2
                sleep 10
            fi
        fi
        set -x
        podman run --rm -ti --security-opt label=disable --privileged                                    \
                   --uidmap=1000:0:1 --uidmap=0:1:1000 --uidmap 1001:1001:64536                          \
                   -v ${PWD}:/srv/ --device /dev/kvm --device /dev/fuse                                  \
                   --tmpfs /tmp -v /var/tmp:/var/tmp --name cosa                                         \
                   ${COREOS_ASSEMBLER_CONFIG_GIT:+-v $COREOS_ASSEMBLER_CONFIG_GIT:/srv/src/config/:ro}   \
                   ${COREOS_ASSEMBLER_GIT:+-v $COREOS_ASSEMBLER_GIT/src/:/usr/lib/coreos-assembler/:ro}  \
                   ${COREOS_ASSEMBLER_CONTAINER_RUNTIME_ARGS}                                            \
                   ${COREOS_ASSEMBLER_CONTAINER:-$COREOS_ASSEMBLER_CONTAINER_LATEST} "$@"
        rc=$?; set +x; return $rc
    }
    ```

* Run the builder

    ``` shell
    cosa shell
    ```

* Initialize the build

    ``` shell
    cosa init https://github.com/coreos/fedora-coreos-config
    ```

* Fetch metadata and packages

    ``` shell
    cosa fetch
    ```

* Build a qemu VM image

    ``` shell
    cosa build
    ```

    Each build will create a new directory in `$PWD/builds/`, containing the generated OSTree commit and the qemu VM image.

* Run the image

    ``` shell
    cosa run
    ```

## Customization

The CoreOS Assembler offers three main customization options:
* [`manifest.yaml`](https://coreos.github.io/coreos-assembler/working/#manifestyaml)

    An rpm-ostree "manifest" or "treefile", primarily, a list of RPMs and their associated repositories.
    See the rpm-ostree documentation for the [treefile format reference](https://coreos.github.io/rpm-ostree/treefile/)

* [`overlay.d/`](https://coreos.github.io/coreos-assembler/working/#overlayd)

    A generic way to embed architecture-independent configuration and scripts by creating subdirectories in `overlay.d/`.
    Each subdirectory is added to the OSTree commit in lexicographic order.

* [`image.yaml`](https://coreos.github.io/coreos-assembler/working/#imageyaml)

    Configuration for the output disk images

Additionally, one may use [`overrides`](https://coreos.github.io/coreos-assembler/working/#using-overrides) to embed local RPMs from the build environment, that should not be pulled from a remote repository:

1. Package the binary as an RPM

2. Add any dependencies of the RPM to `manifest.yaml`

3. Run `cosa fetch` to prepare dependencies

4. Place the RPM in `overrides/rpm`

5. Add the name of your RPM to `manifest.yaml`

6. Run `cosa build`. Your RPM will be added to the final image.


Example: We want to build FCOS with our own kernel

1. Follow [Kernel Building](#kernel-building) to build the kernel

    You should end up with at least three RPMs: `kernel`, `kernel-core`, `kernel-modules`.
    `kernel` depends on `core` and `modules`, `modules` on `core`, and `core` on common FCOS packages (`bash`, `systemd`, etc.).
    These dependencies should already be in the manifest.

2. Run `cosa fetch`

3. Place the kernel RPMs in `overrides/rpm`

    `kernel`, `kernel-core`, `kernel-modules` should already be in the manifest (`src/config/manifests/bootable-rpm-ostree.yaml`)

4. Run `cosa build` to create the image

5. Test the image with `cosa run`

6. Run `cosa buildextend-gcp` and `cosa buildextend-azure` to additionaly create a VM image for GCP and Azure

## RPM packaging

If we want to make the most use of CoreOS assembler we should package our applications as RPM packages.
See [creating rpm packages](https://docs.fedoraproject.org/en-US/quick-docs/creating-rpm-packages/).

Brief overview of the required steps:

1. Create a directory with your source code or binary file

2. Add a <package>.spec file

    Run the following command to create a spec file template that you can update with information about your package
    ``` shell
    rpmdev-newspec <package>
    ```

3. Create the RPM

    ``` shell
    fedpkg --release f35 local
    ```

## Kernel Building

See the [building a custom kernel](https://docs.fedoraproject.org/en-US/quick-docs/kernel/build-custom-kernel/) from the Fedora Project documentation.

The following assumes you are running on a current release of Fedora.
We have a Fedora 35 image available on GCP, make sure you have enough space available and the VM is capable to build the kernel in a reasonable time (e2-standard-8 takes ~2h to finish the build).

1. Install dependencies and clone the kernel

    ``` shell
    sudo dnf install fedpkg fedora-packager rpmdevtools ncurses-devel pesign grubby qt3-devel libXi-devel gcc-c++
    fedpkg clone -a kernel && cd kernel
    sudo dnf builddep kernel.spec
    ```

    Optionally install `ccache` to speed up rebuilds
    ``` shell
    sudo dnf install ccache
    ```

2. Check out the kernel branch you want to base your build on

    Each release has its own branch. E.g. to customize the kernel for Fedora 35, check out `origin/f35`. `rawhide` tracks the latest iteration, following closely behind the mainline kernel.
    ``` shell
    git checkout origin/f35
    git checkout -b custom-kernel
    ```

3. Customize buildid by chaning `# define buildid .local` to `%define buildid .<your_custom_id_here>` in `kernel.spec`

4. Apply your changes and patches to the kernel

5. Build the RPMs

    This will take a while
    ``` shell
    fedpkg local
    ```
    The built kernel RPMs will be in `./x86_64/`

6. You can now use and install the kernel packages

    ``` shell
    sudo dnf install --nogpgcheck ./x86_64/kernel-$version.rpm
    ```
