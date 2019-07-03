from aiopg import connect

import settings


async def _connect():
    return await connect(
        database=settings.POSTGRES_DB,
        user=settings.POSTGRES_USER,
        password=settings.POSTGRES_PASSWORD,
        host=settings.POSTGRES_HOST,
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
    fields = ", ".join(settings.PARTNER_FIELDS)
    sql = f"SELECT {fields} FROM socio WHERE cnpj = '{cnpj}'"
    rows = await _query(connection, sql)

    if not rows:
        return

    return tuple(dict(zip(settings.PARTNER_FIELDS, row)) for row in rows)


async def get_company(cnpj):
    connection = await _connect()
    sql = f"""
        SELECT empresa.*, cnae.descricao AS cnae_fiscal_descricao
        FROM empresa
        INNER JOIN cnae ON empresa.cnae_fiscal = cnae.codigo
        WHERE cnpj = '{cnpj}'
    """

    company = await _query(connection, sql, unique=True)
    if not company:
        return None

    data = dict(zip(settings.COMPANY_FIELDS, company))
    data["cnaes_secundarios"] = await _secondary_activities(connection, cnpj)
    data["qsa"] = await _partners(connection, cnpj)

    connection.close()
    return data
