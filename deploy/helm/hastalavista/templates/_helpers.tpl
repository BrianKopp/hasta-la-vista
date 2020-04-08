{{/* vim: set filetype=mustache: */}}

{{/*
Creates a default fully qualified app name
Truncate to 63 chars because K8s name fields limit.
*/}}
{{- define "fullname" -}}
{{- printf "%s" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
