{{- define "app.name" -}}
{{ .Values.nameOverride | default .Chart.Name }}
{{- end -}}

{{- define "app.labels" -}}
app.kubernetes.io/name: {{ include "app.name" . }}
app.kubernetes.io/part-of: platform-engineering-demo
team: {{ .Values.owner }}
{{- end -}}
