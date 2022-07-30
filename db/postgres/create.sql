CREATE UNLOGGED TABLE IF NOT EXISTS {{ .TableFullName }} (
    {{ .IDFieldName }}       bigint NOT NULL PRIMARY KEY,
    {{ .JSONFieldName }}     jsonb NOT NULL
);
