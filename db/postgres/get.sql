SELECT {{ .JSONFieldName }}
FROM {{ .CompanyTableFullName }}
WHERE id = $1;
