#!/bin/sh

set -eu

dev=vpn_wg0

ip link add dev "${dev}" type wireguard
wg setconf "${dev}" /etc/wireguard/wg.conf
ip link set dev "${dev}" up

for cidr in ${VPN_PEER_CIDRS}
do
    ip route replace "${cidr}" dev "${dev}"
done
