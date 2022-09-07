SELECT {{ .ValueFieldName }}
FROM {{ .MetaTableFullName }}
WHERE key = ?;
