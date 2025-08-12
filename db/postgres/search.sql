SELECT {{ .CursorFieldName }}, {{ .JSONFieldName }}
FROM {{ .CompanyTableFullName }}
WHERE
  {{ if .Query.Cursor -}}
  cursor > {{ .Query.CursorAsInt }} AND
  {{- end }}
  {{ if .Query.UF -}}
  ({{ range $i, $uf := .Query.UF }}{{ if $i }} OR {{ end }}json -> 'uf' = '"{{ $uf }}"'::jsonb{{ end }})
  {{- end }}
  {{ if and .Query.UF .Query.CodigoNaturezaJuridica }} AND {{ end }}
  {{ if .Query.CodigoNaturezaJuridica -}}
  ({{ range $i, $nat:= .Query.CodigoNaturezaJuridica }}{{ if $i }} OR {{ end }}json -> 'codigo_natureza_juridica' = '{{ $nat }}'::jsonb{{ end }})
  {{- end }}
  {{ if and (or .Query.UF .Query.CodigoNaturezaJuridica) .Query.CNAEFiscal }} AND {{ end }}
  {{ if .Query.CNAEFiscal -}}
  ({{ range $i, $cnae := .Query.CNAEFiscal }}{{ if $i }} OR {{ end }}json -> 'cnae_fiscal' = '{{ $cnae }}'::jsonb{{ end }})
  {{- end }}
  {{ if and (or .Query.UF .Query.CodigoNaturezaJuridica .Query.CNAEFiscal) .Query.CNAE }} AND {{ end }}
  {{ if .Query.CNAE -}}
  (
    jsonb_path_query_array(json, '$.cnaes_secundarios[*].codigo') @> '[{{ range $i, $cnae := .Query.CNAE }}{{ if $i }},{{ end }}{{ $cnae }}{{ end }}]'
    {{ range $i, $cnae := .Query.CNAE -}}
    OR json -> 'cnae_fiscal' = '{{ $cnae }}'::jsonb
    {{ end -}}
  )
  {{- end }}
  {{ if and (or .Query.UF .Query.CodigoNaturezaJuridica .Query.CNAEFiscal .Query.CNAE) .Query.CNPF }} AND {{ end }}
  {{ if .Query.CNPF -}}
  (
    jsonb_path_query_array(json, '$.qsa[*].cnpj_cpf_do_socio') @> '[{{ range $i, $cnpf := .Query.CNPF }}{{ if $i }},{{ end }}"{{ $cnpf }}"{{ end }}]'
  )
  {{- end }}
ORDER BY {{ .CursorFieldName }}
LIMIT {{ .Query.Limit }}
