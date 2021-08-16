CREATE TABLE IF NOT EXISTS cnaes (
    codigo int8 NOT NULL,
    descricao text NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_cnaes_codigo ON cnaes USING btree (codigo);
