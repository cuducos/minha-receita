UPDATE {{ .CompanyTableFullName }}
SET {{ .JSONFieldName }} = {{ .CompanyTableName }}.{{ .JSONFieldName }} || ?
WHERE {{ .IDFieldName }} >= ? AND {{ .IDFieldName }} <= ?;
