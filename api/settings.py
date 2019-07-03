from decouple import config


SANIC_HOST = config("SANIC_HOST", default="0.0.0.0")
SANIC_PORT = config("SANIC_PORT", default="8000", cast=int)
SANIC_DEBUG = config("SANIC_DEBUG", default="False", cast=bool)

POSTGRES_HOST = config("POSTGRES_HOST")
POSTGRES_DB = config("POSTGRES_DB")
POSTGRES_USER = config("POSTGRES_USER")
POSTGRES_PASSWORD = config("POSTGRES_PASSWORD")

COMPANY_FIELDS = (
    "cnpj",
    "identificador_matriz_filial",
    "razao_social",
    "nome_fantasia",
    "situacao_cadastral",
    "data_situacao_cadastral",
    "motivo_situacao_cadastral",
    "nome_cidade_exterior",
    "codigo_natureza_juridica",
    "data_inicio_atividade",
    "cnae_fiscal",
    "descricao_tipo_logradouro",
    "logradouro",
    "numero",
    "complemento",
    "bairro",
    "cep",
    "uf",
    "codigo_municipio",
    "municipio",
    "ddd_telefone_1",
    "ddd_telefone_2",
    "ddd_fax",
    "qualificacao_do_responsavel",
    "capital_social",
    "porte",
    "opcao_pelo_simples",
    "data_opcao_pelo_simples",
    "data_exclusao_do_simples",
    "opcao_pelo_mei",
    "situacao_especial",
    "data_situacao_especial",
)

PARTNER_FIELDS = (
    "identificador_de_socio",
    "nome_socio",
    "cnpj_cpf_do_socio",
    "codigo_qualificacao_socio",
    "percentual_capital_social",
    "data_entrada_sociedade",
    "cpf_representante_legal",
    "nome_representante_legal",
    "codigo_qualificacao_representante_legal",
)
