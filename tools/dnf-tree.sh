#!/usr/bin/env bash

set -exuo pipefail
shopt -s inherit_errexit

bazel run //:bazeldnf -- fetch \
  --repofile rpm/repo.yaml
bazel run //:bazeldnf -- rpmtree \
  --workspace WORKSPACE.bazel \
  --to_macro rpm/rpms.bzl%rpms \
  --buildfile rpm/BUILD.bazel \
  --repofile rpm/repo.yaml \
  --name cryptsetup-devel \
  cryptsetup-devel
bazel run //:bazeldnf -- rpmtree \
  --workspace WORKSPACE.bazel \
  --to_macro rpm/rpms.bzl%rpms \
  --buildfile rpm/BUILD.bazel \
  --repofile rpm/repo.yaml \
  --name glibc \
  glibc
bazel run //:bazeldnf -- rpmtree \
  --workspace WORKSPACE.bazel \
  --to_macro rpm/rpms.bzl%rpms \
  --buildfile rpm/BUILD.bazel \
  --repofile rpm/repo.yaml \
  --name libvirt-devel \
  libvirt-devel
bazel run //:bazeldnf -- rpmtree \
  --workspace WORKSPACE.bazel \
  --to_macro rpm/rpms.bzl%rpms \
  --buildfile rpm/BUILD.bazel \
  --repofile rpm/repo.yaml \
  --name containerized-libvirt \
  libvirt-daemon-config-network \
  libvirt-daemon-kvm \
  qemu-kvm \
  swtpm \
  swtpm-tools \
  iptables-legacy \
  dnsmasq \
  libvirt-client
bazel run //:bazeldnf -- prune \
  --workspace WORKSPACE.bazel \
  --to_macro rpm/rpms.bzl%rpms \
  --buildfile rpm/BUILD.bazel
bazel run //rpm:ldd-cryptsetup
bazel run //rpm:ldd-libvirt
bazel run //rpm:ldd-glibc
