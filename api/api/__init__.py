from re import sub

from sanic import Sanic
from sanic.exceptions import NotFound

from api.db import connect
from api.json_tools import json_response
from api.search import company


app = Sanic(__file__)


@app.route("/", methods=("POST",))
async def search(request):
    cnpj = request.form.get("cnpj")
    if not cnpj:
        raise NotFound("CNPJ não enviado na requisição POST.")

    cleaned = sub(r"\D", "", cnpj)
    if not cleaned or len(cleaned) != 14:
        raise NotFound(f"CNPJ {cnpj} inválido.")

    async with connect() as connection:
        obj = await company(connection, cleaned)

    if not obj:
        raise NotFound(f"CNPJ {cnpj} não encontrado.")

    return json_response(obj)
