{{/*
Controller labels
*/}}
{{- define "tugboat-controller.labels" -}}
{{ include "tugboat-controller.selectorLabels" . }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "tugboat-controller.selectorLabels" -}}
app.kubernetes.io/component: controller
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "tugboat-controller.serviceAccountName" -}}
  {{- if .Values.tugboatController.serviceAccount.create }}
    {{- default (printf "%s-serviceaccount" (include "tugboat.fullname" .)) .Values.tugboatController.serviceAccount.name }}
  {{- else }}
    {{- default "default" .Values.tugboatController.serviceAccount.name }}
  {{- end }}
{{- end }}
