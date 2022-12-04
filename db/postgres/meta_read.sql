SELECT {{ .ValueFieldName }}
FROM {{ .MetaTableFullName }}
WHERE key = $1;
