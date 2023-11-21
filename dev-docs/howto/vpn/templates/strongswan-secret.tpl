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
              local_ts  = {{ .Values.podCIDR }},{{ .Values.serviceCIDR }}
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
