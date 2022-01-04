SELECT {{ .JSONFieldName }}
FROM {{ .TableFullName }}
WHERE {{ .IDFieldName }} LIKE ?;
