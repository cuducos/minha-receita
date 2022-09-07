SELECT {{ .JSONFieldName }}
FROM {{ .CompanyTableFullName }}
WHERE id = ?;
