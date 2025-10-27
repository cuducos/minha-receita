CREATE TABLE IF NOT EXISTS {{ .CompanyTableFullName }} (
    {{ .CursorFieldName }} SERIAL PRIMARY KEY,
    {{ .IDFieldName }} char(14) NOT NULL,
    {{ .JSONFieldName }} jsonb NOT NULL
);
CREATE TABLE IF NOT EXISTS {{ .MetaTableFullName }} (
    {{ .KeyFieldName }} char(16) NOT NULL PRIMARY KEY,
    {{ .ValueFieldName }} text NOT NULL
);
CREATE UNIQUE INDEX {{ .CompanyTableName }}_id ON {{ .CompanyTableFullName }} ({{ .IDFieldName }});
CREATE UNIQUE INDEX {{ .MetaTableName }}_{{ .KeyFieldName }} ON {{ .MetaTableFullName }} ({{ .KeyFieldName }});
