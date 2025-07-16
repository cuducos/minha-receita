{{ $tableName := .CompanyTableFullName }}
{{ $jsonField := .JSONFieldName }}
{{range .ExtraIndexes }}
    {{ if .IsRoot }}
        CREATE INDEX IF NOT EXISTS "idx_{{ .Name }}" ON {{ $tableName }} USING BTREE (({{ $jsonField }}->'{{ .Value }}'));
    {{ else }}
        CREATE INDEX IF NOT EXISTS "idx_{{ .Name }}" ON {{ $tableName }} USING GIN (
            (jsonb_path_query_array({{ $jsonField }}, '{{ .NestedPath }}'))
        );
    {{ end }}
{{ end }}
