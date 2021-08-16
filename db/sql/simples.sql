CREATE TABLE IF NOT EXISTS simples (
    cnpj char(8) NOT NULL,
    opcao_pelo_simples varchar(1) NULL,
    data_opcao_pelo_simples date NULL,
    data_exclusao_do_simples date NULL,
    opcao_pelo_mei varchar(1) NULL,
    data_opcao_pelo_mei date NULL,
    data_entrada_do_mei date NULL
);

CREATE INDEX IF NOT EXISTS idx_simples_cnpj ON simples USING btree (cnpj);
