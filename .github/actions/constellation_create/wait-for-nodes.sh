#!/bin/bash

# We don't want to abort the script if there's a transient error in kubectl.
set +e 
set -uo pipefail

NODES_COUNT=$(( CONTROL_NODES_COUNT + WORKER_NODES_COUNT ))
JOINWAIT=0

# Reports how many nodes are registered and fulfill condition=ready.
num_nodes_ready() {
    kubectl get nodes -o json | 
      jq '.items | map(select(.status.conditions[] | .type == "Ready" and .status == "True")) | length'
}

# Reports how many API server pods are ready.
num_apiservers_ready() {
    kubectl get pods -n kube-system -l component=kube-apiserver -o json |
      jq '.items | map(select(.status.conditions[] | .type == "Ready" and .status == "True")) | length'
}

# Prints node joining progress.
report_join_progress() {
    echo -n "nodes_joined=$(kubectl get nodes -o json | jq '.items | length')/${NODES_COUNT} "
    echo -n "nodes_ready=$(num_nodes_ready)/${NODES_COUNT} "
    echo "api_servers_ready=$(num_apiservers_ready)/${CONTROL_NODES_COUNT} ..."
}

# Indicates by exit code whether the cluster is ready, i.e. all nodes and API servers are ready.
cluster_ready() {
    [[ "$(num_nodes_ready)" == "${NODES_COUNT}" ]] && [[ "$(num_apiservers_ready)" == "${CONTROL_NODES_COUNT}" ]]
}

echo "::group::Wait for nodes"
until cluster_ready || [[ ${JOINWAIT} -gt ${JOINTIMEOUT} ]];
do
    report_join_progress
    JOINWAIT=$((JOINWAIT+30))
    sleep 30
done
report_join_progress
if [[ ${JOINWAIT} -gt ${JOINTIMEOUT} ]]; then
    set -x
    kubectl get nodes -o wide
    kubectl get pods -n kube-system -o wide
    kubectl get events -n kube-system
    set +x
    echo "::error::timeout reached before all nodes became ready"
    echo "::endgroup::"
    exit 1
fi
echo "::endgroup::"