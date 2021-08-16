CREATE TABLE IF NOT EXISTS municipios (
    codigo int8 NOT NULL,
    descricao text NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_municipios_codigo ON municipios USING btree (codigo);
