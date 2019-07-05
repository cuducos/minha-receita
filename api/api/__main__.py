from api import app
from api.settings import SANIC_HOST, SANIC_PORT, SANIC_DEBUG


app.run(host=SANIC_HOST, port=SANIC_PORT, debug=SANIC_DEBUG)
