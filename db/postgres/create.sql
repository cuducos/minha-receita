CREATE TABLE IF NOT EXISTS {{ .TableFullName }} (
    {{ .IDFieldName }}       char(14) NOT NULL PRIMARY KEY,
    {{ .BaseCNPJFieldName }} char(8) NOT NULL,
    {{ .JSONFieldName }}     json NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_table_{{ .TableName }}_on_column_{{ .BaseCNPJFieldName }}
    ON {{ .TableFullName }} USING btree ({{ .BaseCNPJFieldName }});
