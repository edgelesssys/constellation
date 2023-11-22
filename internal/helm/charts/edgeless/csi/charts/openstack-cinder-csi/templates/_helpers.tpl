{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "cinder-csi.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "cinder-csi.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "cinder-csi.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "cinder-csi.labels" -}}
app.kubernetes.io/name: {{ include "cinder-csi.name" . }}
helm.sh/chart: {{ include "cinder-csi.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}


{{/*
Create the name of the service account to use
*/}}
{{- define "cinder-csi.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "cinder-csi.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Create unified labels for cinder-csi components
*/}}
{{- define "cinder-csi.common.matchLabels" -}}
app: {{ template "cinder-csi.name" . }}
release: {{ .Release.Name }}
{{- end -}}

{{- define "cinder-csi.common.metaLabels" -}}
chart: {{ template "cinder-csi.chart" . }}
heritage: {{ .Release.Service }}
{{- if .Values.extraLabels }}
{{ toYaml .Values.extraLabels -}}
{{- end }}
{{- end -}}

{{- define "cinder-csi.controllerplugin.matchLabels" -}}
component: controllerplugin
{{ include "cinder-csi.common.matchLabels" . }}
{{- end -}}

{{- define "cinder-csi.controllerplugin.labels" -}}
{{ include "cinder-csi.controllerplugin.matchLabels" . }}
{{ include "cinder-csi.common.metaLabels" . }}
{{- end -}}

{{- define "cinder-csi.nodeplugin.matchLabels" -}}
component: nodeplugin
{{ include "cinder-csi.common.matchLabels" . }}
{{- end -}}

{{- define "cinder-csi.nodeplugin.labels" -}}
{{ include "cinder-csi.nodeplugin.matchLabels" . }}
{{ include "cinder-csi.common.metaLabels" . }}
{{- end -}}

{{- define "cinder-csi.snapshot-controller.matchLabels" -}}
component: snapshot-controller
{{ include "cinder-csi.common.matchLabels" . }}
{{- end -}}

{{- define "cinder-csi.snapshot-controller.labels" -}}
{{ include "cinder-csi.snapshot-controller.matchLabels" . }}
{{ include "cinder-csi.common.metaLabels" . }}
{{- end -}}
