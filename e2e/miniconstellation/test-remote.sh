#!/usr/bin/env bash
#
# This script expects:
# Constellation CLI @ $PWD/constellation
# kubectl @ PATH

set -euo pipefail

debug_info() {
  arg=$?

  echo "Getting debug info"

  sha256sum ./*.raw

  ls -lisah

  cat ./constellation-conf.yaml

  exit "${arg}"
}

trap debug_info ERR

echo "::group::Run E2E Test"
mkdir constellation_workspace
cd constellation_workspace
cp ../constellation .
chmod u+x constellation

# wait for docker to come up
echo "Waiting for docker service to be active..."

# Wait at most 10min
count=0
until systemctl is-active docker || [[ ${count} -eq 60 ]]; do
  sleep 10
  count=$((count + 1))
done

if [[ ${count} -eq 60 ]]; then
  echo "Docker service did not come up in time."
  exit 1
fi

echo "Done waiting."

./constellation mini up --debug

export KUBECONFIG="${PWD}/constellation-admin.conf"

# Wait for nodes to actually show up in K8s
count=0
until kubectl wait --for=condition=Ready --timeout=2s nodes control-plane-0 2> /dev/null || [[ ${count} -eq 30 ]]; do
  echo "Control-planes are not registered in Kubernetes yet. Waiting..."
  sleep 10
  count=$((count + 1))
done

count=0
until kubectl wait --for=condition=Ready --timeout=2s nodes worker-0 2> /dev/null || [[ ${count} -eq 30 ]]; do
  echo "Worker nodes are not registered in Kubernetes yet. Waiting..."
  sleep 10
  count=$((count + 1))
done

# Wait for deployments
kubectl -n kube-system wait --for=condition=Available=True --timeout=180s deployment coredns
kubectl -n kube-system wait --for=condition=Available=True --timeout=180s deployment cilium-operator
# Wait for daemon sets
kubectl -n kube-system rollout status --timeout 180s daemonset cilium
kubectl -n kube-system rollout status --timeout 180s daemonset join-service
kubectl -n kube-system rollout status --timeout 180s daemonset key-service
kubectl -n kube-system rollout status --timeout 180s daemonset konnectivity-agent
kubectl -n kube-system rollout status --timeout 180s daemonset verification-service

echo "Miniconstellation started successfully. Shutting down..."

./constellation mini down -y
echo "::endgroup::"
