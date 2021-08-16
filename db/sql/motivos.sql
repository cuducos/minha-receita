CREATE TABLE IF NOT EXISTS motivos_situacao_cadastral (
    codigo int8 NOT NULL,
    descricao text NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_motivos_situacao_cadastral_codigo ON motivos_situacao_cadastral USING btree (codigo);
