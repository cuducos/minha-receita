CREATE TABLE IF NOT EXISTS naturezas_juridicas (
    codigo int8 NOT NULL,
    descricao text NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_naturezas_juridicas_codigo ON naturezas_juridicas USING btree (codigo);
