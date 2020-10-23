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
