
{{- define "..name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 42 | trimSuffix "-" }}
{{- end }}

{{- define "..fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 42 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 42 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{- define "..chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 42 | trimSuffix "-" }}
{{- end }}

{{- define "..labels" -}}
helm.sh/chart: {{ include "..chart" . }}
{{ include "..selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "..selectorLabels" -}}
app.kubernetes.io/name: {{ include "..name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "..commonEnv" -}}
- name: VPN_PEER_CIDRS
  value: {{ join " " .Values.peerCIDRs | quote }}
- name: VPN_POD_CIDR
  value: {{ .Values.podCIDR | quote }}
- name: VPN_SERVICE_CIDR
  value: {{ .Values.serviceCIDR | quote }}
{{- end }}
