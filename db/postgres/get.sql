SELECT {{ .IDFieldName }}, {{ .JSONFieldName }}
FROM {{ .TableFullName }}
WHERE id = ?;
