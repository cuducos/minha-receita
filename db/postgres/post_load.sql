CREATE INDEX idx_remove_duplicates ON {{ .CompanyTableFullName }} ({{ .IDFieldName }});

DELETE FROM {{ .CompanyTableFullName }}
WHERE ctid IN (
  SELECT ctid
  FROM (
    SELECT
      ctid,
      row_number() OVER (
        PARTITION BY ({{ .IDFieldName }})
        ORDER BY ctid DESC
      ) AS count
    FROM {{ .CompanyTableFullName }}
  ) t
  WHERE count > 1
);

DROP INDEX idx_remove_duplicates;
ALTER TABLE {{ .CompanyTableFullName }} DROP COLUMN tmp_pk CASCADE;
CREATE UNIQUE INDEX {{ .CompanyTableName }}_pk ON {{ .CompanyTableFullName }} ({{ .IDFieldName }});
ALTER TABLE cnpj ADD PRIMARY KEY USING INDEX {{ .CompanyTableName }}_pk;
ALTER TABLE {{ .CompanyTableFullName }} SET LOGGED;
