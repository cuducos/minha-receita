from asynctest import patch

from api import app


def test_get_returns_405():
    _, response = app.test_client.get("/")
    assert response.status == 405
    assert response.text == "Error: Method GET not allowed for URL /"


def test_post_returns_404_if_sent_without_data():
    _, response = app.test_client.post("/", data={})
    assert response.status == 404
    assert response.text == "Error: CNPJ não enviado na requisição POST."


def test_post_returns_404_if_sent_without_cnpj():
    _, response = app.test_client.post("/", data={"cpf": "00.000.000/0000-00"})
    assert response.status == 404
    assert response.text == "Error: CNPJ não enviado na requisição POST."


def test_post_returns_404_for_non_existent_cnpj():
    with patch("api.company") as mock:
        mock.return_value = None
        _, response = app.test_client.post("/", data={"cnpj": "00.000.000/0000-00"})

    assert response.status == 404
    assert response.text == "Error: CNPJ 00.000.000/0000-00 não encontrado."


def test_post_returns_200_for_existing_cnpj_masked():
    with patch("api.company") as mock:
        mock.return_value = {"hell": "yeah"}
        _, response = app.test_client.post("/", data={"cnpj": "19.131.243/0001-97"})

    assert response.status == 200
    assert response.json == {"hell": "yeah"}


def test_post_returns_200_for_existing_cnpj_unmasked():
    with patch("api.company") as mock:
        mock.return_value = {"hell": "yeah"}
        _, response = app.test_client.post("/", data={"cnpj": "19131243000197"})

    assert response.status == 200
    assert response.json == {"hell": "yeah"}
