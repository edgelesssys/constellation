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

# Wait at most 20min
count=0
until systemctl is-active docker || [[ ${count} -eq 120 ]]; do
  sleep 10
  count=$((count + 1))
done

if [[ ${count} -eq 120 ]]; then
  echo "Docker service did not come up in time."
  exit 1
fi

echo "Done waiting."

./constellation mini up --debug

export KUBECONFIG="${PWD}/constellation-admin.conf"

# Wait for nodes to actually show up in K8s (taken from .github/actions/constellation_create/action.yml)
echo "::group::Wait for nodes"
NODES_COUNT=2
JOINWAIT=0
JOINTIMEOUT="600" # 10 minutes timeout for all nodes to join
until [[ "$(kubectl get nodes -o json | jq '.items | length')" == "${NODES_COUNT}" ]] || [[ $JOINWAIT -gt $JOINTIMEOUT ]]; do
  echo "$(kubectl get nodes -o json | jq '.items | length')/${NODES_COUNT} nodes have joined.. waiting.."
  JOINWAIT=$((JOINWAIT + 30))
  sleep 30
done
if [[ $JOINWAIT -gt $JOINTIMEOUT ]]; then
  echo "Timed out waiting for nodes to join"
  exit 1
fi
echo "$(kubectl get nodes -o json | jq '.items | length')/${NODES_COUNT} nodes have joined"
if ! kubectl wait --for=condition=ready --all nodes --timeout=20m; then
  kubectl get pods -n kube-system
  kubectl get events -n kube-system
  echo "::error::kubectl wait timed out before all nodes became ready"
  echo "::endgroup::"
  exit 1
fi
echo "::endgroup::"

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
