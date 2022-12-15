UPDATE {{ .CompanyTableFullName }}
SET {{ .JSONFieldName }} = jsonb_set(
    {{ .CompanyTableFullName }}.{{ .JSONFieldName }},
    array['{{ .PartnersJSONFieldName }}'],
    CASE
        WHEN {{ .CompanyTableFullName }}.{{ .JSONFieldName }}->'{{ .PartnersJSONFieldName }}'::text = 'null'  THEN $3
        ELSE {{ .CompanyTableFullName }}.{{ .JSONFieldName }}->'{{ .PartnersJSONFieldName }}' || $3
    END,
    false
)
WHERE {{ .IDFieldName }} >= $1 AND {{ .IDFieldName }} <= $2;
