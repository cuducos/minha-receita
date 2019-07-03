from re import sub

from sanic import Sanic
from sanic.exceptions import NotFound
from sanic.response import json

from search import get_company
from settings import SANIC_HOST, SANIC_PORT, SANIC_DEBUG


app = Sanic(__file__)


@app.route("/", methods=("POST",))
async def search(request):
    cnpj = request.form.get("cnpj")
    if not cnpj:
        raise NotFound("CNPJ não enviado na requisição POST.")

    cleaned = sub(r"\D", "", cnpj)
    if not cleaned or len(cleaned) != 14:
        raise NotFound(f"CNPJ {cnpj} inválido.")

    company = await get_company(cleaned)
    if not company:
        raise NotFound(f"CNPJ {cnpj} não encontrado.")

    return json(company)


if __name__ == "__main__":
    app.run(host=SANIC_HOST, port=SANIC_PORT, debug=SANIC_DEBUG)
