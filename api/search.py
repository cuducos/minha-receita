from aiopg import connect
from decouple import config


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


async def _query(connection, sql, unique=False):
    cursor = await connection.cursor()
    await cursor.execute(sql)

    if unique:
        return await cursor.fetchone()

    return await cursor.fetchall()


async def _secondary_activities(connection, cnpj):
    sql = f"""
        SELECT cnae_secundaria.cnae, cnae.descricao
        FROM cnae_secundaria
        INNER JOIN cnae ON cnae_secundaria.cnae = cnae.codigo
        WHERE cnpj = '{cnpj}'
    """

    rows = await _query(connection, sql)
    if not rows:
        return

    fields = ("cnae_codigo", "cnae_descricao")
    return tuple(dict(zip(fields, row)) for row in rows)


async def _partners(connection, cnpj):
    fields = ", ".join(PARTNER_FIELDS)
    sql = f"SELECT {fields} FROM socio WHERE cnpj = '{cnpj}'"
    rows = await _query(connection, sql)

    if not rows:
        return

    return tuple(dict(zip(PARTNER_FIELDS, row)) for row in rows)


async def get_company(cnpj):
    connection = await connect(
        database=config("POSTGRES_DB"),
        user=config("POSTGRES_USER"),
        password=config("POSTGRES_PASSWORD"),
        host=config("POSTGRES_HOST"),
    )

    sql = f"""
        SELECT empresa.*, cnae.descricao AS cnae_fiscal_descricao
        FROM empresa
        INNER JOIN cnae ON empresa.cnae_fiscal = cnae.codigo
        WHERE cnpj = '{cnpj}'
    """

    company = await _query(connection, sql, unique=True)
    if not company:
        return None

    data = dict(zip(COMPANY_FIELDS, company))
    data["cnaes_secundarios"] = await _secondary_activities(connection, cnpj)
    data["qsa"] = await _partners(connection, cnpj)

    connection.close()
    return data
