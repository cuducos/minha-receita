{{ $tableName := .CompanyTableFullName }}
{{range .ExtraIndexes }}
    {{ if .IsRoot }}
        CREATE INDEX IF NOT EXISTS "idx_{{ .Name }}" ON {{ $tableName }} USING BTREE ((json->'{{ .Value }}'));
    {{ else }}
        CREATE INDEX IF NOT EXISTS "idx_{{ .Name }}" ON {{ $tableName }} USING BTREE (jsonb_extract_path(json,'{{ .Value }}') jsonb_ops);
    {{ end }}
{{ end }}
