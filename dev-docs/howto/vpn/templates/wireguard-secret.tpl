{{- define "wireguard.conf" }}
[Interface]
ListenPort = {{ .Values.wireguard.port }}
PrivateKey = {{ .Values.wireguard.private_key }}
[Peer]
PublicKey = {{ .Values.wireguard.peer_key }}
AllowedIPs = {{ join "," .Values.peerCIDRs }}
{{- if .Values.wireguard.endpoint }}
Endpoint = {{- .Values.wireguard.endpoint }}
{{- end }}
{{- if .Values.wireguard.keepAlive }}
PersistentKeepalive = {{- .Values.wireguard.keepAlive }}
{{- end }}
{{ end }}