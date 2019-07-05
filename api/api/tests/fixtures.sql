--
-- empresa
--

CREATE TABLE empresa (
    cnpj text,
    identificador_matriz_filial bigint,
    razao_social text,
    nome_fantasia text,
    situacao_cadastral bigint,
    data_situacao_cadastral date,
    motivo_situacao_cadastral bigint,
    nome_cidade_exterior text,
    codigo_natureza_juridica bigint,
    data_inicio_atividade date,
    cnae_fiscal bigint,
    descricao_tipo_logradouro text,
    logradouro text,
    numero text,
    complemento text,
    bairro text,
    cep bigint,
    uf text,
    codigo_municipio bigint,
    municipio text,
    ddd_telefone_1 text,
    ddd_telefone_2 text,
    ddd_fax text,
    qualificacao_do_responsavel bigint,
    capital_social numeric,
    porte bigint,
    opcao_pelo_simples boolean,
    data_opcao_pelo_simples text,
    data_exclusao_do_simples text,
    opcao_pelo_mei boolean,
    situacao_especial text,
    data_situacao_especial text
);

INSERT INTO empresa (cnpj, identificador_matriz_filial, razao_social, nome_fantasia, situacao_cadastral, data_situacao_cadastral, motivo_situacao_cadastral, nome_cidade_exterior, codigo_natureza_juridica, data_inicio_atividade, cnae_fiscal, descricao_tipo_logradouro, logradouro, numero, complemento, bairro, cep, uf, codigo_municipio, municipio, ddd_telefone_1, ddd_telefone_2, ddd_fax, qualificacao_do_responsavel, capital_social, porte, opcao_pelo_simples, data_opcao_pelo_simples, data_exclusao_do_simples, opcao_pelo_mei, situacao_especial, data_situacao_especial) VALUES ('19131243000197', 1, 'OPEN KNOWLEDGE BRASIL', 'REDE PELO CONHECIMENTO LIVRE', 2, '2013-10-03', 0, NULL, 3999, '2013-10-03', 9430800, 'ALAMEDA', 'FRANCA', '144', 'APT   34', 'JARDIM PAULISTA', 1422000, 'SP', 7107, 'SAO PAULO', '11  23851939', NULL, NULL, 10, 0.00, 5, false, NULL, NULL, false, NULL, NULL);

--
-- socio
--

CREATE TABLE socio (
    cnpj text,
    identificador_de_socio bigint,
    nome_socio text,
    cnpj_cpf_do_socio text,
    codigo_qualificacao_socio bigint,
    percentual_capital_social bigint,
    data_entrada_sociedade date,
    cpf_representante_legal text,
    nome_representante_legal text,
    codigo_qualificacao_representante_legal bigint
);

INSERT INTO socio (cnpj, identificador_de_socio, nome_socio, cnpj_cpf_do_socio, codigo_qualificacao_socio, percentual_capital_social, data_entrada_sociedade, cpf_representante_legal, nome_representante_legal, codigo_qualificacao_representante_legal) VALUES ('19131243000197', 2, 'NATALIA PASSOS MAZOTTE CORTEZ', '***059967**', 10, 0, '2019-02-14', NULL, NULL, NULL);

--
-- cnae_secundaria
--

CREATE TABLE cnae_secundaria (
    cnpj text,
    cnae bigint
);

INSERT INTO cnae_secundaria (cnpj, cnae) VALUES ('19131243000197', 9493600);
INSERT INTO cnae_secundaria (cnpj, cnae) VALUES ('19131243000197', 9499500);
INSERT INTO cnae_secundaria (cnpj, cnae) VALUES ('19131243000197', 8599699);
INSERT INTO cnae_secundaria (cnpj, cnae) VALUES ('19131243000197', 8230001);
INSERT INTO cnae_secundaria (cnpj, cnae) VALUES ('19131243000197', 6204000);

--
-- cnae
--

CREATE TABLE cnae (
    codigo bigint,
    descricao text
);

INSERT INTO cnae (codigo, descricao) VALUES (9430800, 'Atividades de associações de defesa de direitos sociais');
INSERT INTO cnae (codigo, descricao) VALUES (6204000, 'Consultoria em tecnologia da informação');
INSERT INTO cnae (codigo, descricao) VALUES (8230001, 'Serviços de organização de feiras, congressos, exposições e festas');
INSERT INTO cnae (codigo, descricao) VALUES (8599699, 'Outras atividades de ensino não especificadas anteriormente');
INSERT INTO cnae (codigo, descricao) VALUES (9493600, 'Atividades de organizações associativas ligadas à cultura e à arte');
INSERT INTO cnae (codigo, descricao) VALUES (9499500, 'Atividades associativas não especificadas anteriormente');
