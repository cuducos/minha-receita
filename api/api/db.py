from contextlib import asynccontextmanager

import aiopg
from psycopg2.extras import RealDictCursor

from api import settings


@asynccontextmanager
async def connect(custom_settings=None):
    _settings = custom_settings or settings
    dsn = (
        f"dbname={_settings.POSTGRES_DB} "
        f"user={_settings.POSTGRES_USER} "
        f"password={_settings.POSTGRES_PASSWORD} "
        f"host={_settings.POSTGRES_HOST} "
        f"port={_settings.POSTGRES_PORT}"
    )
    pool = await aiopg.create_pool(dsn)
    yield pool
    pool.close()


async def query(pool, sql, unique=False):
    with (await pool.cursor(cursor_factory=RealDictCursor)) as cursor:
        await cursor.execute(sql)

        if unique:
            return await cursor.fetchone()

        return await cursor.fetchall()
