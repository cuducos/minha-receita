{{- if eq .ExtraIndexesFields.Type "root" }}
CREATE INDEX IF NOT EXISTS "idx_{{ .ExtraIndexesFields.Name }}" ON {{ .CompanyTableFullName }} USING GIN ((json->'{{ .ExtraIndexesFields.Value }}'));
{{- else }}
CREATE INDEX IF NOT EXISTS "idx_{{ .ExtraIndexesFields.Name }}" ON {{ .CompanyTableFullName }} USING GIN (jsonb_extract_path(json, '{{ .ExtraIndexesFields.Type }}', '{{ .ExtraIndexesFields.Value }}') jsonb_ops);
{{- end }}