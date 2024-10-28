CREATE TABLE IF NOT EXISTS {{ .CompanyTableFullName }} (
    tmp_pk SERIAL PRIMARY KEY,
    {{ .IDFieldName }} char(14) NOT NULL,
    {{ .JSONFieldName }} jsonb NOT NULL
);
CREATE TABLE IF NOT EXISTS {{ .MetaTableFullName }} (
    {{ .KeyFieldName }} char(16) NOT NULL PRIMARY KEY,
    {{ .ValueFieldName }} text NOT NULL
)
