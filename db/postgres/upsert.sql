INSERT INTO {{ .TableFullName }} ({{ .IDFieldName }}, {{ .JSONFieldName }})
VALUES(?, ?)
ON CONFLICT ({{ .IDFieldName }}) DO
UPDATE SET {{ .JSONFieldName }} = ?;
