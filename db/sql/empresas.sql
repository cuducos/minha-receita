CREATE TABLE IF NOT EXISTS empresas (
    cnpj char(8) NOT NULL,
    razao_social text NULL,
    natureza_juridica integer NULL,
    qualificacao_do_responsavel integer NULL,
    capital_social decimal NULL,
    porte integer NULL,
    ente_federativo_resposavel text NULL
);

CREATE INDEX IF NOT EXISTS idx_empresas_cnpj ON empresas USING btree (cnpj);
