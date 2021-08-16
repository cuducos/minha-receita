CREATE TABLE IF NOT EXISTS socios (
    cnpj char(8) NOT NULL,
	identificador int8 NULL,
	nome_razao_social text NULL,
	cpf_cnpj text NULL,
	qualificacao int8 NULL,
	data_entrada date NULL,
    pais text NULL,
	cpf_representante_legal text NULL,
	nome_representante_legal text NULL,
	qualificacao_representante_legal int8 NULL,
    faixa_etaria text NULL
);

CREATE INDEX IF NOT EXISTS idx_socios_cnpj ON socios USING btree (cnpj);

-- TODO: mask first 3 digits and last 2 digits of the CPF

-- TODO: faixa etária
-- 1 para os intervalos entre 0 a 12 anos;
-- 2 para os intervalos entre 13 a 20 anos;
-- 3 para os intervalos entre 21 a 30 anos;
-- 4 para os intervalos entre 31 a 40 anos;
-- 5 para os intervalos entre 41 a 50 anos;
-- 6 para os intervalos entre 51 a 60 anos;
-- 7 para os intervalos entre 61 a 70 anos;
-- 8 para os intervalos entre 71 a 80 anos;
-- 9 para maiores de 80 anos;
-- 0 para não se aplica.
