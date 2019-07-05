from api.db import query
from api.settings import COMPANY_FIELDS, PARTNER_FIELDS


async def get_secondary_activities(connection, cnpj):
    sql = f"""
        SELECT cnae_secundaria.cnae, cnae.descricao
        FROM cnae_secundaria
        INNER JOIN cnae ON cnae_secundaria.cnae = cnae.codigo
        WHERE cnpj = '{cnpj}'
    """

    rows = await query(connection, sql)
    if not rows:
        return

    fields = ("cnae_codigo", "cnae_descricao")
    return tuple(dict(zip(fields, row)) for row in rows)


async def get_partners(connection, cnpj):
    fields = ", ".join(PARTNER_FIELDS)
    sql = f"SELECT {fields} FROM socio WHERE cnpj = '{cnpj}'"
    rows = await query(connection, sql)

    if not rows:
        return

    return tuple(dict(zip(PARTNER_FIELDS, row)) for row in rows)


async def get_company(connection, cnpj):
    sql = f"""
        SELECT empresa.*, cnae.descricao AS cnae_fiscal_descricao
        FROM empresa
        LEFT JOIN cnae ON empresa.cnae_fiscal = cnae.codigo
        WHERE cnpj = '{cnpj}'
    """

    company = await query(connection, sql, unique=True)
    if not company:
        return None

    data = dict(zip(COMPANY_FIELDS, company))
    data["cnaes_secundarios"] = await get_secondary_activities(connection, cnpj)
    data["qsa"] = await get_partners(connection, cnpj)
    return data
