{{ define "common.podSpec.containerPortsAndProbes" }}
ports:
- name: http
  containerPort: {{ .Values.application.serverPort }}
  protocol: TCP
- name: health-check
  containerPort: {{ .Values.application.managementPort }}
  protocol: TCP
{{- if .Values.application.startupProbe.enabled }}
startupProbe:
  {{- include "common.podSpec.probeTemplate" .Values.application.startupProbe }}
{{- end }}
{{- if .Values.application.liveness.enabled }}
livenessProbe:
  {{- include "common.podSpec.probeTemplate" .Values.application.liveness }}
{{- end }}
{{- if .Values.application.readiness.enabled }}
readinessProbe:
  {{- include "common.podSpec.probeTemplate" .Values.application.readiness }}
  successThreshold: {{ .Values.application.readiness.successThreshold }}
{{- end }}
{{- end }}

{{ define "common.podSpec.probeTemplate" }}
  {{- $probeType := .type -}}
  {{- $valid := list "httpGet" "tcpSocket" "exec" }}
  {{- if not (has $probeType $valid) }}
  {{- fail "Invalid probe type, must be one of (httpGet,tcpSocket,exec)" }}
  {{- end }}
  {{ $probeType }}:
    {{- if eq $probeType "httpGet" }}
    path: {{ .path }}
    port: {{ .port }}
    host: {{ .host }}{{/*     Host name to connect to, defaults to the pod IP. You probably want to set "Host" in httpHeaders instead.*/}}
    httpHeaders: {{ toYaml .httpHeaders | nindent 6 }}
    scheme: {{ .scheme }}
    {{- end }}
    {{- if eq $probeType "tcpSocket" }}
    port: {{ .port }}
    host: {{ .host }}{{/*     Host name to connect to, defaults to the pod IP. You probably want to set "Host" in httpHeaders instead.*/}}
    {{- end }}
    {{- if eq $probeType "exec" }}
    command: {{ toYaml .command | nindent 6 }}
    {{- end }}
  periodSeconds: {{ .periodSeconds }}
  timeoutSeconds: {{ .timeoutSeconds }}
  failureThreshold: {{ .failureThreshold }}
  initialDelaySeconds: {{ .initialDelaySeconds }}
  timeoutSeconds: {{ .timeoutSeconds }}
{{- end }}

{{ define "common.podSpec.mainPart" }}
{{- with .Values.imagePullSecrets -}}
imagePullSecrets:
{{- toYaml . | nindent 0 }}
{{- end }}
serviceAccountName: {{ default "default" .Values.global.serviceAccountName }}
terminationGracePeriodSeconds: {{ .Values.application.terminationGracePeriodSeconds }}
volumes:
{{- if .Values.extraVolumes }}
{{- tpl .Values.extraVolumes . | trim | nindent 0 }}
{{- end }}
initContainers:
{{- if .Values.extraInitContainers }}
{{- tpl .Values.extraInitContainers . | trim | nindent 0 }}
{{- end }}
containers:
- name: {{ .Chart.Name }}
  image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
  imagePullPolicy: {{ .Values.image.pullPolicy }}
  {{- if .Values.application.command }}
  command:
    {{- toYaml .Values.application.command | nindent 4 }}
  {{- end }}
  {{- if .Values.application.args }}
  args:
    {{- toYaml .Values.application.args | nindent 4 }}
  {{- end }}
  env:
  {{- include "helpers.list-env-variables" . | nindent 2 }}
  lifecycle:
  {{- if .Values.application.lifecycle }}
    {{- toYaml .Values.application.lifecycle | nindent 4 }}
  {{- end }}
  volumeMounts:
    {{- if .Values.extraVolumeMounts }}
    {{- tpl .Values.extraVolumeMounts . | trim | nindent 4 }}
    {{- end }}
  resources: {{- toYaml .Values.resources | nindent 4 }}
{{- end -}}

{{ define "common.podSpec.selectorsTolerationsAffinity" }}
{{- with .Values.nodeSelector }}
nodeSelector: {{- toYaml . | nindent 2 }}
{{- end }}
{{- with .Values.affinity }}
affinity: {{- toYaml . | nindent 2 }}
{{- end }}
{{- with .Values.tolerations }}
tolerations: {{- toYaml . | nindent 2 }}
{{- end }}
{{- end -}}

{{ define "common.podAnnotations" }}
{{- if .Values.metrics.enabled }}
prometheus.io/scrape: {{ .Values.metrics.enabled | quote }}
prometheus.io/port: {{ .Values.metrics.port | quote }}
prometheus.io/path: {{ .Values.metrics.path | quote }}
{{- end }}
{{- if .Values.defaultIpPool }}
cni.projectcalico.org/ipv4pools: '["default-pool"]'
{{- end }}
{{- if .Values.env.vault }}
vault.security.banzaicloud.io/vault-addr: {{ .Values.global.vaultAddress | quote }}
vault.security.banzaicloud.io/vault-role: {{ .Release.Namespace | quote }}
{{- end }}
{{- range $key, $value := .Values.podAnnotations }}
{{ $key }}: {{ $value | quote }}
{{- end }}
{{- end -}}
