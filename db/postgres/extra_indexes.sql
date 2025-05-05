{{ $tableName := .CompanyTableFullName }}
{{range .ExtraIndexes }}
    {{ if .IsRoot }}
        CREATE INDEX IF NOT EXISTS "idx_{{ .Name }}" ON {{ $tableName }} USING GIN ((json->'{{ .Value }}'));
    {{ else }}
        CREATE INDEX IF NOT EXISTS "idx_{{ .Name }}" ON {{ $tableName }} USING GIN (jsonb_extract_path(json,'{{ .Value }}') jsonb_ops);
    {{ end }}
{{ end }}
