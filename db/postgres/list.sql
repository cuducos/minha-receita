SELECT {{ .JSONFieldName }}
FROM {{ .TableFullName }}
WHERE {{ .BaseCNPJFieldName }} = ?;
