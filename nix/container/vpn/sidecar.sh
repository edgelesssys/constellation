#!/bin/sh

set -u

if [ "$$" -eq "1" ]; then
  echo 'This script must run in the root PID namespace, but $$ == 1!' >&2
  exit 1
fi

myip() {
  ip -j addr show eth0 | jq -r '.[0].addr_info[] | select(.family == "inet") | .local'
}

# Disable source IP verification on our network interface. Otherwise, VPN
# packets will be dropped by Cilium.
reconcile_sip_verification() {
  # We want all of the cilium calls in this function to target the same
  # process, so that we fail if the agent restarts in between. Thus, we only
  # query the pid once per reconciliation.
  cilium_agent=$(pidof cilium-agent) || return 0

  cilium() {
    nsenter -t "${cilium_agent}" -a -r -w cilium "$@"
  }

  myendpoint=$(cilium endpoint get "ipv4:$(myip)" | jq '.[0].id') || return 0

  if [ "$(cilium endpoint config "${myendpoint}" -o json | jq -r .realized.options.SourceIPVerification)" = "Enabled" ]; then
    cilium endpoint config "${myendpoint}" SourceIPVerification=Disabled
  fi
}

optional_mtu() {
  if [ -n "${VPN_MTU}" ]; then
    printf "mtu %s" "${VPN_MTU}"
  fi
}

# Set up the route from the node network namespace to the VPN pod.
reconcile_route() {
  for cidr in ${VPN_PEER_CIDRS}; do
    # shellcheck disable=SC2046 # Word splitting is intentional here.
    nsenter -t 1 -n ip route replace "${cidr}" via "$(myip)" $(optional_mtu)
  done
}

while true; do
  reconcile_route
  reconcile_sip_verification
  sleep 10
done
