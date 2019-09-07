from decouple import config


SANIC_HOST = config("SANIC_HOST", default="0.0.0.0")
SANIC_PORT = config("SANIC_PORT", default="8000", cast=int)
SANIC_DEBUG = config("SANIC_DEBUG", default="False", cast=bool)

POSTGRES_DB = config("POSTGRES_DB")
POSTGRES_HOST = config("POSTGRES_HOST")
POSTGRES_PASSWORD = config("POSTGRES_PASSWORD")
POSTGRES_PORT = config("POSTGRES_PORT", cast=int)
POSTGRES_USER = config("POSTGRES_USER")
