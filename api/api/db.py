from contextlib import asynccontextmanager

import aiopg

from api import settings


@asynccontextmanager
async def connect(custom_settings=None):
    _settings = custom_settings or settings
    connection = await aiopg.connect(
        database=_settings.POSTGRES_DB,
        user=_settings.POSTGRES_USER,
        password=_settings.POSTGRES_PASSWORD,
        host=_settings.POSTGRES_HOST,
    )
    yield connection
    connection.close()


async def query(connection, sql, unique=False):
    cursor = await connection.cursor()
    await cursor.execute(sql)

    if unique:
        return await cursor.fetchone()

    return await cursor.fetchall()
