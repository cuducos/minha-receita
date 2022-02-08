UPDATE {{ .TableFullName }}
SET {{ .JSONFieldName }} = {{ .TableName }}.{{ .JSONFieldName }} || ?
WHERE {{ .IDFieldName }} >= ? AND {{ .IDFieldName }} <= ?;
