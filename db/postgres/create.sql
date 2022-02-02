CREATE TABLE IF NOT EXISTS {{ .TableFullName }} (
    {{ .IDFieldName }}       bigint NOT NULL PRIMARY KEY,
    {{ .BaseCNPJFieldName }} integer NOT NULL,
    {{ .JSONFieldName }}     jsonb NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_table_{{ .TableName }}_on_column_{{ .BaseCNPJFieldName }}
    ON {{ .TableFullName }} USING btree ({{ .BaseCNPJFieldName }});
