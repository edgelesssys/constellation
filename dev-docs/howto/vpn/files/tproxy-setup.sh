#!/bin/sh

set -eu

### Pod IPs ###

# Pod IPs are just NATed.

iptables -t nat -N VPN_POST || iptables -t nat -F VPN_POST

for cidr in ${VPN_PEER_CIDRS}
do
    iptables -t nat -A VPN_POST -s "${cidr}" -d "${VPN_POD_CIDR}" -j MASQUERADE
done

iptables -t nat -C POSTROUTING -j VPN_POST || iptables -t nat -A POSTROUTING -j VPN_POST

### Service IPs ###

# Service IPs need to be connected to locally to trigger the cgroup connect hook, thus we send them to the transparent proxy.

# Packets with mark 1 are for tproxy and need to be delivered locally.
pref=42
table=42
mark=0x1/0x1
ip rule add pref "${pref}" fwmark "${mark}" lookup "${table}"
ip route replace local 0.0.0.0/0 dev lo table "${table}"

iptables -t mangle -N VPN_PRE || iptables -t mangle -F VPN_PRE

for cidr in ${VPN_PEER_CIDRS}
do
    for proto in tcp udp
    do
        iptables -t mangle -A VPN_PRE -p "${proto}" -s "${cidr}" -d "${VPN_SERVICE_CIDR}" \
          -j TPROXY --tproxy-mark "${mark}" --on-port 61001
    done
done

iptables -t mangle -C PREROUTING -j VPN_PRE || iptables -t mangle -A PREROUTING -j VPN_PRE
