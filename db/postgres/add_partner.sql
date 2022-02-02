UPDATE {{ .TableFullName }}
SET {{ .JSONFieldName }} = jsonb_set(
    {{ .TableFullName }}.{{ .JSONFieldName }},
    '{qsa}',
    CASE
        WHEN {{ .TableFullName }}.{{ .JSONFieldName }}->'qsa'::text = 'null'  THEN ?
        ELSE {{ .TableFullName }}.{{ .JSONFieldName }}->'qsa' || ?
    END,
    false
)
WHERE {{ .BaseCNPJFieldName }} = ?;
