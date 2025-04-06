{{- if eq .OptExtraIndexes.TypeIdx "root" }}
CREATE INDEX IF NOT EXISTS idx_{{ .OptExtraIndexes.NameIdx }} ON {{ .CompanyTableFullName }} USING GIN ((json->'{{ .OptExtraIndexes.ValueIdx }}'));
{{- else if ne .OptExtraIndexes.TypeIdx "root" }}
CREATE INDEX IF NOT EXISTS idx_{{ .OptExtraIndexes.NameIdx }} ON {{ .CompanyTableFullName }} USING GIN (jsonb_extract_path(json, '{{ .OptExtraIndexes.TypeIdx }}', '{{ .OptExtraIndexes.ValueIdx }}') jsonb_ops);
{{- else }}

{{- end }}