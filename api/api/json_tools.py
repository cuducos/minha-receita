import json
from datetime import date
from decimal import Decimal

from sanic import response


class JsonEncoder(json.JSONEncoder):
    """Supports date and decimal objects"""

    def default(self, value):
        if isinstance(value, date):
            return value.strftime("%Y-%m-%d")
        if isinstance(value, Decimal):
            return float(value)
        return super(JsonEncoder, self).default(value)


def json_response(data):
    return response.json(data, dumps=json.dumps, cls=JsonEncoder, ensure_ascii=False)
