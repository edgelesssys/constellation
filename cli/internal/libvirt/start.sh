#!/bin/bash

# Assign qemu the GID of the host system's 'kvm' group to avoid permission issues for environments defaulting to 660 for /dev/kvm (e.g. Debian-based distros)
KVM_HOST_GID="$(stat -c '%g' /dev/kvm)"
groupadd -o -g "$KVM_HOST_GID" host-kvm
usermod -a -G host-kvm qemu

# Start libvirt daemon
libvirtd --daemon --listen
virtlogd --daemon

sleep infinity
