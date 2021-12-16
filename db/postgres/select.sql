SELECT *
FROM {{ .TableFullName }}
WHERE id = ?;
