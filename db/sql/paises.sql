CREATE TABLE IF NOT EXISTS paises (
    codigo int8 NOT NULL,
    descricao text NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_paises_codigo ON paises USING btree (codigo);
