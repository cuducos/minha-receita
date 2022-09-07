INSERT INTO {{ .MetaTableName }} ({{ .KeyFieldName }}, {{ .ValueFieldName }})
VALUES (?, ?)
ON CONFLICT ({{ .KeyFieldName }})
DO UPDATE
SET {{ .ValueFieldName }} = ?
