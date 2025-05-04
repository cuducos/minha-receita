{{range .ExtraIndexesFields }}
    {{ if .IsRoot }}
        CREATE INDEX IF NOT EXISTS "idx_{{ .Name }}" ON {{ .CompanyTableFullName }} USING GIN ((json->'{{ .Value }}'));
    {{ else }}
        CREATE INDEX IF NOT EXISTS "idx_{{ .Name }}" ON {{ .CompanyTableFullName }} USING GIN (jsonb_extract_path(json,'{{ .Value }}') jsonb_ops);
    {{ end }}
{{ end }}
