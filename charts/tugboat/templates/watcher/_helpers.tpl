{{/*
Create the name of the service account to use
*/}}
{{- define "tugboat-watcher.serviceAccountName" -}}
  {{- if .Values.tugboatWatcher.serviceAccount.create }}
    {{- default (include "tugboat.fullname" .) .Values.tugboatWatcher.serviceAccount.name }}
  {{- else }}
    {{- default "default" .Values.tugboatWatcher.serviceAccount.name }}
  {{- end }}
{{- end }}

{{/*
Controller labels
*/}}
{{- define "tugboat-watcher.labels" -}}
{{ include "tugboat-watcher.selectorLabels" . }}
{{- end }}

{{- define "tugboat-watcher.selectorLabels" -}}
app.kubernetes.io/component: watcher
{{- end }}

