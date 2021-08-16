CREATE TABLE IF NOT EXISTS qualificacoes_de_socios (
    codigo int8 NOT NULL,
    descricao text NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_qualificacoes_de_socios_codigo ON qualificacoes_de_socios USING btree (codigo);
