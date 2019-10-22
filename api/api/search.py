from asyncio import gather

from api.db import query


def remove_key(ban, dictionary):
    return {key: value for key, value in dictionary.items() if key != ban}


async def secondary_activities(connection, cnpj):
    sql = f"""
        SELECT cnae_secundaria.cnae AS codigo, cnae.descricao
        FROM cnae_secundaria
        INNER JOIN cnae ON cnae_secundaria.cnae = cnae.codigo
        WHERE cnpj = '{cnpj}'
    """

    rows = await query(connection, sql)
    if not rows:
        return

    # convert psycopg2's RealDict to normal dict
    return tuple({key: value for key, value in row.items()} for row in rows)


async def partners(connection, cnpj):
    sql = f"SELECT * FROM socio WHERE cnpj = '{cnpj}'"
    rows = await query(connection, sql)
    return tuple(remove_key("cnpj", row) for row in rows) if rows else None


async def company(connection, cnpj):
    sql = f"""
        SELECT empresa.*, cnae.descricao AS cnae_fiscal_descricao
        FROM empresa
        LEFT JOIN cnae ON empresa.cnae_fiscal = cnae.codigo
        WHERE cnpj = '{cnpj}'
    """

    row = await query(connection, sql, unique=True)
    if not row:
        return None

    row["cnaes_secundarios"], row["qsa"] = await gather(
        secondary_activities(connection, cnpj), partners(connection, cnpj)
    )
    return row
