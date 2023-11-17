
{{- define "tproxy.script" }}

iptables_add() {
    if ! iptables -C "$@"
    then
        iptables -A "$@"
    fi
}

### Pod IPs ###

# Pod IPs are just NATed.

{{- range .Values.peerCIDRs }}
iptables_add POSTROUTING -t nat -s {{ . }} -d {{ $.Values.podIDR | default "10.10.0.0/16" }} -j MASQUERADE
{{- end }} 

### Service IPs ###

# Service IPs need to be connected to locally to trigger the cgroup connect hook, thus we send them to the transparent proxy.

# Packets with mark 1 are for tproxy and need to be delivered locally.
ip rule add pref 42 fwmark 1 lookup 42
ip route replace local 0.0.0.0/0 dev lo table 42

{{- range .Values.peerCIDRs }}
iptables_add PREROUTING -t mangle -p tcp -s {{ . }} -d {{ $.Values.serviceCIDR | default "10.96.0.0/12" }} -j TPROXY --tproxy-mark 0x1/0x1 --on-port 61001
iptables_add PREROUTING -t mangle -p udp -s {{ . }} -d {{ $.Values.serviceCIDR | default "10.96.0.0/12"  }} -j TPROXY --tproxy-mark 0x1/0x1 --on-port 61001
{{- end }} 

{{- end }}