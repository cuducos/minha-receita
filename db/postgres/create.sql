CREATE TABLE IF NOT EXISTS {{ .TableFullName }} (
   {{ .IDFieldName }}   char(14) NOT NULL PRIMARY KEY,
   {{ .JSONFieldName }} json NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_table_{{ .TableName }}_on_column_{{ .IDFieldName }}
    ON {{ .TableFullName }} USING btree ({{ .IDFieldName }});
