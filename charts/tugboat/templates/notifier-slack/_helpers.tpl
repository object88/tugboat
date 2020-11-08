{{/*
Create the name of the service account to use
*/}}
{{- define "tugboat-notifier-slack.serviceAccountName" -}}
  {{- if .Values.tugboatNotifierSlack.serviceAccount.create }}
    {{- default (printf "%s-notifier-slack" (include "tugboat.fullname" .)) .Values.tugboatNotifierSlack.serviceAccount.name }}
  {{- else }}
    {{- default "default" .Values.tugboatNotifierSlack.serviceAccount.name }}
  {{- end }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "tugboat-notifier-slack.labels" -}}
{{ include "tugboat-notifier-slack.selectorLabels" . }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "tugboat-notifier-slack.selectorLabels" -}}
app.kubernetes.io/component: notifier-slack
{{- end }}
