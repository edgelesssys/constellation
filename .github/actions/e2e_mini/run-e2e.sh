#!/usr/bin/env bash
#
# This script installs all dependencies for MiniConstellation to be run on a
# fresh Ubuntu 22.04 LTS installation.
# It expects to find the to be used Constellation CLI to be available at
# $HOME/constellation
#
set -euxo pipefail

echo "::group::Install dependencies"
cloud-init status --wait

export DEBIAN_FRONTEND=noninteractive
sudo apt update -y

sudo apt install -y bridge-utils cpu-checker \
  libvirt-clients libvirt-daemon libvirt-daemon-system \
  qemu qemu-kvm virtinst xsltproc \
  ca-certificates curl gnupg lsb-release

sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt update -y
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
sudo usermod -aG docker "$USER"
newgrp docker
echo "::endgroup::"

echo "::group::Run E2E Test"
mkdir constellation_workspace
cd constellation_workspace
mv "$HOME"/constellation .
chmod u+x constellation

sudo sh -c 'echo "127.0.0.1 license.confidential.cloud" >> /etc/hosts'

mkdir -p "$HOME"/.docker
touch "$HOME"/.docker/config.json

./constellation mini up

curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install kubectl /usr/local/bin/kubectl

export KUBECONFIG="$PWD/constellation-admin.conf"

# Wait for nodes to actually show up in K8s control plane
sleep 10

# Wait for nodes
kubectl wait --for=condition=Ready --timeout=600s nodes control-plane-0
kubectl wait --for=condition=Ready --timeout=600s nodes worker-0
# Wait for deployments
kubectl -n kube-system wait --for=condition=Available=True --timeout=180s deployment coredns
kubectl -n kube-system wait --for=condition=Available=True --timeout=180s deployment cilium-operator
# Wait for daemon sets
kubectl -n kube-system rollout status --timeout 180s daemonset cilium
kubectl -n kube-system rollout status --timeout 180s daemonset join-service
kubectl -n kube-system rollout status --timeout 180s daemonset kms
kubectl -n kube-system rollout status --timeout 180s daemonset konnectivity-agent
kubectl -n kube-system rollout status --timeout 180s daemonset verification-service
echo "::endgroup::"
