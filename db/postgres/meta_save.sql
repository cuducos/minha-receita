INSERT INTO {{ .MetaTableName }} ({{ .KeyFieldName }}, {{ .ValueFieldName }})
VALUES ($1, $2)
ON CONFLICT ({{ .KeyFieldName }})
DO UPDATE
SET {{ .ValueFieldName }} = $2
