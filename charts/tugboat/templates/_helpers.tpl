{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "tugboat.chart" -}}
  {{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Expand the name of the chart.
*/}}
{{- define "tugboat.name" -}}
  {{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "tugboat.fullname" -}}
  {{- if .Values.fullnameOverride }}
    {{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
  {{- else }}
    {{- $name := default .Chart.Name .Values.nameOverride }}
    {{- if contains $name .Release.Name }}
      {{- .Release.Name | trunc 63 | trimSuffix "-" }}
    {{- else }}
      {{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
    {{- end }}
  {{- end }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "tugboat.labels" -}}
helm.sh/chart: {{ include "tugboat.chart" . }}
{{ include "tugboat.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "tugboat.selectorLabels" -}}
app.kubernetes.io/name: {{ include "tugboat.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}


{{- define "image.tag" -}}
  {{- include "nested-get" (dict "root" . "args" (list "image" "tag")) | default .Chart.AppVersion -}}
{{- end }}

{{- define "image.pullPolicy" -}}
  {{- include "nested-get" (dict "root" . "args" (list "image" "pullPolicy")) | default "IfNotPresent" -}}
{{- end }}

{{- define "nested-get" -}}
  {{- $vdict := (dict "root" .root "args" (concat (list "Values") .args)) -}}
  {{- $vgdict := (dict "root" .root "args" (concat (list "Values" "global") .args)) -}}
  {{- include "inner-nested-get" $vdict | default (include "inner-nested-get" $vgdict) -}}
{{- end -}}

{{- define "inner-nested-get" -}}
  {{- if empty .root -}}
    {{- "" -}}
  {{- else if (empty .args) -}}
    {{- .root -}}
  {{- else if (hasKey .root (first .args)) -}}
    {{ include "inner-nested-get" (dict "root" (get .root (first .args)) "args" (rest .args)) }}
  {{- else -}}
    {{- "" -}}
  {{- end -}}
{{- end -}}
