{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "helm-common.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "helm-common.fullname" -}}
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
{{- define "helm-common.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "helm-common.labels" -}}
helm.sh/chart: {{ include "helm-common.chart" . }}
{{ include "helm-common.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "helm-common.selectorLabels" -}}
app.kubernetes.io/name: {{ include "helm-common.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "helpers.list-env-variables" }}
{{- if .Values.env }}
{{- if .Values.appEnvSecret }}
{{- $appSecretName := .Values.appEnvSecret.name -}}
{{- range $key, $val := .Values.env.secret }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ $appSecretName }}
      key: {{ $key }}
{{- end }}
{{- end }}
{{- if .Values.appEnvConfigMap }}
{{- $appConfigmapName := .Values.appEnvConfigMap.name -}}
{{- range $key, $val := .Values.env.configMap }}
- name: {{ $key }}
  valueFrom:
    configMapKeyRef:
      name: {{ $appConfigmapName }}
      key: {{ $key }}
{{- end }}
{{- end }}
{{- range $key, $val := .Values.env.normal }}
- name: {{ $key }}
  value: {{ tpl ( $val | quote ) $ }}
{{- end }}
{{- range $key, $val := .Values.env.vault }}
- name: {{ $key }}
  value: vault:k8s/data/{{ $.Release.Namespace }}/{{ $val }}
{{- end }}
{{- end }}
{{- end }}
