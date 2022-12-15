UPDATE {{ .CompanyTableFullName }}
SET {{ .JSONFieldName }} = {{ .CompanyTableName }}.{{ .JSONFieldName }} || $3
WHERE {{ .IDFieldName }} >= $1 AND {{ .IDFieldName }} <= $2;
