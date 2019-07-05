import json
from datetime import date
from decimal import Decimal

from api.json_tools import JsonEncoder, json_response


FIXTURE = {"date": date(2019, 2, 14), "pi": Decimal("3.1415")}


def test_json_encoder():
    converted = json.dumps(FIXTURE, cls=JsonEncoder)
    assert converted == '{"date": "2019-02-14", "pi": 3.1415}'


def test_json_response():
    response = json_response(FIXTURE)
    assert response.status == 200
    assert response.content_type == "application/json"
    assert response.body == b'{"date": "2019-02-14", "pi": 3.1415}'
