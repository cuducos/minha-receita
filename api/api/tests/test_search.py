from datetime import date

import pytest

from api.search import get_company, get_partners, get_secondary_activities


EXPECTED = {
    "cnpj": "19131243000197",
    "identificador_matriz_filial": 1,
    "razao_social": "OPEN KNOWLEDGE BRASIL",
    "nome_fantasia": "REDE PELO CONHECIMENTO LIVRE",
    "situacao_cadastral": 2,
    "data_situacao_cadastral": date(2013, 10, 3),
    "motivo_situacao_cadastral": 0,
    "nome_cidade_exterior": None,
    "codigo_natureza_juridica": 3999,
    "data_inicio_atividade": date(2013, 10, 3),
    "cnae_fiscal": 9430800,
    "descricao_tipo_logradouro": "ALAMEDA",
    "logradouro": "FRANCA",
    "numero": "144",
    "complemento": "APT   34",
    "bairro": "JARDIM PAULISTA",
    "cep": 1422000,
    "uf": "SP",
    "codigo_municipio": 7107,
    "municipio": "SAO PAULO",
    "ddd_telefone_1": "11  23851939",
    "ddd_telefone_2": None,
    "ddd_fax": None,
    "qualificacao_do_responsavel": 10,
    "capital_social": 0.0,
    "porte": 5,
    "opcao_pelo_simples": False,
    "data_opcao_pelo_simples": None,
    "data_exclusao_do_simples": None,
    "opcao_pelo_mei": False,
    "situacao_especial": None,
    "data_situacao_especial": None,
    "cnaes_secundarios": (
        {
            "cnae_codigo": 9493600,
            "cnae_descricao": "Atividades de organizações associativas ligadas à cultura e à arte",
        },
        {
            "cnae_codigo": 9499500,
            "cnae_descricao": "Atividades associativas não especificadas anteriormente",
        },
        {
            "cnae_codigo": 8599699,
            "cnae_descricao": "Outras atividades de ensino não especificadas anteriormente",
        },
        {
            "cnae_codigo": 8230001,
            "cnae_descricao": "Serviços de organização de feiras, congressos, exposições e festas",
        },
        {
            "cnae_codigo": 6204000,
            "cnae_descricao": "Consultoria em tecnologia da informação",
        },
    ),
    "qsa": (
        {
            "identificador_de_socio": 2,
            "nome_socio": "NATALIA PASSOS MAZOTTE CORTEZ",
            "cnpj_cpf_do_socio": "***059967**",
            "codigo_qualificacao_socio": 10,
            "percentual_capital_social": 0,
            "data_entrada_sociedade": date(2019, 2, 14),
            "cpf_representante_legal": None,
            "nome_representante_legal": None,
            "codigo_qualificacao_representante_legal": None,
        },
    ),
}


@pytest.mark.asyncio
async def test_partners_exists(db):
    partners = await get_partners(db, "19131243000197")
    assert len(partners) == len(EXPECTED["qsa"])
    for partner in EXPECTED["qsa"]:
        assert partner in partners


@pytest.mark.asyncio
async def test_partners_does_not_exist(db):
    partners = await get_partners(db, "00000000000000")
    assert partners == None


@pytest.mark.asyncio
async def test_secondary_activities_exists(db):
    secondary_activities = await get_secondary_activities(db, "19131243000197")
    assert len(secondary_activities) == len(EXPECTED["cnaes_secundarios"])
    for activity in EXPECTED["cnaes_secundarios"]:
        assert activity in secondary_activities


@pytest.mark.asyncio
async def test_secondary_activities_does_not_exist(db):
    secondary_activities = await get_secondary_activities(db, "00000000000000")
    assert secondary_activities == None


@pytest.mark.asyncio
async def test_company_exists(db):
    company = await get_company(db, "19131243000197")
    for key in company.keys():
        if key in {"qsa", "cnaes_secundarios"}:
            continue

        assert company[key] == EXPECTED[key]


@pytest.mark.asyncio
async def test_company_does_not_exist(db):
    company = await get_company(db, "00000000000000")
    assert company == None
