SELECT {{ .JSONFieldName }}
FROM {{ .CompanyTableFullName }}
WHERE
  {{ if .Query.Cursor -}}
  id > '{{ .Query.Cursor }}' AND
  {{- end }}
  {{ if .Query.UF -}}
  ({{ range $i, $uf := .Query.UF }}{{ if $i }} OR {{ end }}json -> 'uf' = '"{{ $uf }}"'::jsonb{{ end }})
  {{- end }}
  {{ if and .Query.UF .Query.CNAEFiscal }} AND {{ end }}
  {{ if .Query.CNAEFiscal -}}
  ({{ range $i, $cnae := .Query.CNAEFiscal }}{{ if $i }} OR {{ end }}json -> 'cnae_fiscal' = '{{ $cnae }}'::jsonb{{ end }})
  {{- end }}
  {{ if and (or .Query.UF .Query.CNAEFiscal) .Query.CNAE }} AND {{ end }}
  {{ if .Query.CNAE -}}
  (
    jsonb_path_query_array(json, '$.cnaes_secundarios[*].codigo') @> '[{{ range $i, $cnae := .Query.CNAE }}{{ if $i }},{{ end }}{{ $cnae }}{{ end }}]'
    {{ range $i, $cnae := .Query.CNAE -}}
    OR json -> 'cnae_fiscal' = '{{ $cnae }}'::jsonb
    {{ end -}}
  )
  {{- end }}
LIMIT {{ .Query.Limit }}
