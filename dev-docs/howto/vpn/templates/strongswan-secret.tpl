{{- define "strongswan.swanctl-conf" }}
connections {
    net-net {
    remote_addrs = {{ .Values.ipsec.peer }}
    local {
        auth = psk
    }
    remote {
        auth = psk
    }
    children {
        net-net {
            local_ts  = {{ .Values.podCIDR | default "10.10.0.0/16" }},{{ .Values.serviceCIDR | default "10.96.0.0/12" }}
            remote_ts = {{ join "," .Values.peerCIDRs }}
            start_action = trap
        }
    }
    }
}

secrets {
    ike {
        secret = {{ quote .Values.ipsec.psk }}
    }
}
{{- end }}