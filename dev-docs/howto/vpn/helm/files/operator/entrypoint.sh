#!/bin/sh

signaled() {
  exit 143
}

trap signaled INT TERM

all_ips() {
  kubectl get pods "${VPN_FRONTEND_POD}" -o go-template --template '{{ range .status.podIPs }}{{ printf "%s " .ip }}{{ end }}'
  echo "${VPN_PEER_CIDRS}"
}

cep_patch() {
  for ip in $(all_ips); do printf '{"ipv4": "%s"}' "${ip}"; done | jq -s -c -j |
    jq '[{op: "replace", path: "/status/networking/addressing", value: . }]'
}

# Format the space-separated CIDRs into a JSON array.
vpn_cidrs=$(for ip in ${VPN_PEER_CIDRS}; do printf '"%s" ' "${ip}"; done | jq -s -c -j)

masq_patch() {
  kubectl -n kube-system get configmap ip-masq-agent -o json |
    jq -r .data.config |
    jq "{ masqLinkLocal: .masqLinkLocal, nonMasqueradeCIDRs: ((.nonMasqueradeCIDRs - ${vpn_cidrs}) + ${vpn_cidrs}) }" |
    jq '@json | [{op: "replace", path: "/data/config", value: . }]'
}

reconcile_masq() {
  if ! kubectl -n kube-system get configmap ip-masq-agent > /dev/null; then
    # We don't know enough to create an ip-masq-agent.
    return 0
  fi

  kubectl -n kube-system patch configmap ip-masq-agent --type json --patch "$(masq_patch)" > /dev/null
}

while true; do
  # Reconcile CiliumEndpoint to advertise VPN CIDRs.
  kubectl patch ciliumendpoint "${VPN_FRONTEND_POD}" --type json --patch "$(cep_patch)" > /dev/null

  # Reconcile ip-masq-agent configuration to exclude VPN traffic.
  reconcile_masq

  sleep 10
done
