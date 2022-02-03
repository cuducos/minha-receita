UPDATE {{ .TableFullName }}
SET {{ .JSONFieldName }} = jsonb_set(
    {{ .TableFullName }}.{{ .JSONFieldName }},
    array['{{ .PartnersJSONFieldName }}'],
    CASE
        WHEN {{ .TableFullName }}.{{ .JSONFieldName }}->'{{ .PartnersJSONFieldName }}'::text = 'null'  THEN ?
        ELSE {{ .TableFullName }}.{{ .JSONFieldName }}->'{{ .PartnersJSONFieldName }}' || ?
    END,
    false
)
WHERE {{ .IDFieldName }} >= ? AND {{ .IDFieldName }} <= ?;
