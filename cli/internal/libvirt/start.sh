#!/bin/bash

# Assign qemu the GID of the host system's 'kvm' group to avoid permission issues for environments defaulting to 660 for /dev/kvm (e.g. Debian-based distros)
usermod -a -G $(stat -c '%g' /dev/kvm) qemu

# Start libvirt daemon
libvirtd --daemon --listen
virtlogd --daemon

sleep infinity
