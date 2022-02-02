UPDATE {{ .TableFullName }}
SET {{ .JSONFieldName }} = {{ .TableName }}.{{ .JSONFieldName }} || ?
WHERE {{ .BaseCNPJFieldName }} = ?;
