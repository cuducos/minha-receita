from contextlib import contextmanager
from pathlib import Path
from functools import partial

import psycopg2
from psycopg2.extensions import ISOLATION_LEVEL_AUTOCOMMIT
from psycopg2.errors import DuplicateDatabase
from pytest import fixture

from api import settings
from api.db import connect


FIXTURES = Path() / "api" / "tests" / "fixtures.sql"
ORIGINAL_POSTGRES_DB = settings.POSTGRES_DB
POSTGRES_TEST_DB = f"{settings.POSTGRES_DB}_test"


@contextmanager
def postgres(database, isolation_level=None):
    connection = psycopg2.connect(
        dbname=database,
        user=settings.POSTGRES_USER,
        host=settings.POSTGRES_HOST,
        password=settings.POSTGRES_PASSWORD,
    )
    connection.set_isolation_level(ISOLATION_LEVEL_AUTOCOMMIT)
    yield connection.cursor()
    connection.close()


@fixture(scope="module")
def create_test_db():
    with postgres(ORIGINAL_POSTGRES_DB) as cursor:
        try:
            cursor.execute(f"CREATE DATABASE {POSTGRES_TEST_DB}")
        except DuplicateDatabase:
            cursor.execute(f"DROP DATABASE {POSTGRES_TEST_DB}")
            cursor.execute(f"CREATE DATABASE {POSTGRES_TEST_DB}")

    with postgres(POSTGRES_TEST_DB) as cursor, FIXTURES as fixtures:
        cursor.execute(fixtures.read_text())
        yield cursor

        # block further connections
        cursor.execute(
            f"""
            UPDATE pg_database
            SET datallowconn = 'false'
            WHERE datname = '{POSTGRES_TEST_DB}'
        """
        )

    with postgres(ORIGINAL_POSTGRES_DB) as cursor:
        cursor.execute(
            f"""
            SELECT pg_terminate_backend(pid)
            FROM pg_stat_activity
            WHERE datname = '{POSTGRES_TEST_DB}';
        """
        )
        cursor.execute(f"DROP DATABASE {POSTGRES_TEST_DB}")


@fixture
async def db(create_test_db):
    settings.POSTGRES_DB = POSTGRES_TEST_DB
    async with connect(settings) as connection:
        yield connection
    settings.POSTGRES_DB = ORIGINAL_POSTGRES_DB
