{{- if eq .OptExtraIndexes.Type "root" }}
CREATE INDEX IF NOT EXISTS idx_{{ .OptExtraIndexes.Name }} ON {{ .CompanyTableFullName }} USING GIN ((json->'{{ .OptExtraIndexes.Value }}'));
{{- else }}
CREATE INDEX IF NOT EXISTS idx_{{ .OptExtraIndexes.Name }} ON {{ .CompanyTableFullName }} USING GIN (jsonb_extract_path(json, '{{ .OptExtraIndexes.Type }}', '{{ .OptExtraIndexes.Value }}') jsonb_ops);
{{- end }}