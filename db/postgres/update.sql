UPDATE {{ .TableFullName }}
SET {{ .JSONFieldName }} = ?
WHERE {{ .IDFieldName }} = ?;
