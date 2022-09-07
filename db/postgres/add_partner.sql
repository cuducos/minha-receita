UPDATE {{ .CompanyTableFullName }}
SET {{ .JSONFieldName }} = jsonb_set(
    {{ .CompanyTableFullName }}.{{ .JSONFieldName }},
    array['{{ .PartnersJSONFieldName }}'],
    CASE
        WHEN {{ .CompanyTableFullName }}.{{ .JSONFieldName }}->'{{ .PartnersJSONFieldName }}'::text = 'null'  THEN ?
        ELSE {{ .CompanyTableFullName }}.{{ .JSONFieldName }}->'{{ .PartnersJSONFieldName }}' || ?
    END,
    false
)
WHERE {{ .IDFieldName }} >= ? AND {{ .IDFieldName }} <= ?;
