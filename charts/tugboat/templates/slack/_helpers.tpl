{{/*
Create the name of the service account to use
*/}}
{{- define "tugboat-slack.serviceAccountName" -}}
  {{- if .Values.tugboatSlack.serviceAccount.create }}
    {{- default (printf "%s-slack" (include "tugboat.fullname" .)) .Values.tugboatSlack.serviceAccount.name }}
  {{- else }}
    {{- default "default" .Values.tugboatSlack.serviceAccount.name }}
  {{- end }}
{{- end }}

{{- define "tugboat-slack.labels" -}}
{{ include "tugboat-slack.selectorLabels" . }}
{{- end }}

{{- define "tugboat-slack.selectorLabels" -}}
app.kubernetes.io/component: slack
{{- end }}

