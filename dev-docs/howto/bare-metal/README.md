# Bare-metal SNP setup for Constellation

## Prepare Host

The bare-metal host machine needs to be able to start SEV-SNP VMs.
A thorough explanation can be found here: <https://github.com/AMDESE/AMDSEV/tree/snp-latest>.

First checkout the snp-latest branch:

```bash
git clone https://github.com/AMDESE/AMDSEV.git
cd AMDSEV
git checkout snp-latest
```

Then enable TPM2 support by setting `-DTPM2_ENABLE` in the OVMF build command
found in `common.sh`:

```patch
diff --git a/common.sh b/common.sh
index 9eee947..52bf507 100755
--- a/common.sh
+++ b/common.sh
@@ -155,7 +155,7 @@ build_install_ovmf()
                GCCVERS="GCC5"
        fi

-       BUILD_CMD="nice build -q --cmd-len=64436 -DDEBUG_ON_SERIAL_PORT=TRUE -n $(getconf _NPROCESSORS_ONLN) ${GCCVERS:+-t $GCCVERS} -a X64 -p OvmfPkg/OvmfPkgX64.dsc"
+       BUILD_CMD="nice build -q --cmd-len=64436 -DTPM2_ENABLE -DDEBUG_ON_SERIAL_PORT=TRUE -n $(getconf _NPROCESSORS_ONLN) ${GCCVERS:+-t $GCCVERS} -a X64 -p OvmfPkg/OvmfPkgX64.dsc"

        # initialize git repo, or update existing remote to currently configured one
        if [ -d ovmf ]; then
```

Build and package the binaries. Then install the newly build kernel:

```bash
./build.sh --package
cd linux
dpkg -i linux-image-6.9.0-rc7-snp-host-05b10142ac6a_6.9.0-rc7-g05b10142ac6a-2_amd64.deb
```

Reboot, verify that the right BIOS setting are set as described in
<https://github.com/AMDESE/AMDSEV/tree/snp-latest?tab=readme-ov-file#prepare-host>
and select the new kernel in the boot menu. Note that GRUB usually automatically
select the newest installed kernel as default.

Download a Constellation qemu image, the `constellation-conf.yaml`, and
the `launch-constellation.sh` script in the directory right next to the
`AMDSEV` folder.

```bash
wget https://raw.githubusercontent.com/edgelesssys/constellation/main/dev-docs/howto/bare-metal/launch-constellation.sh
wget https://cdn.confidential.cloud/constellation/v1/ref/main/stream/console/v2.17.0-pre.0.20240516182331-5fb2a2cb89f2/image/csp/qemu/qemu-vtpm/image.raw
wget < link to the constellation CLI provided by Edgeless >
wget < link to the constellation config provided by Edgeless >
```

Install and setup [docker](https://docs.docker.com/engine/install/),
install swtpm, dnsmasq and tmux.

Then simply run:

```bash
sudo ./launch-constellation.sh
```
