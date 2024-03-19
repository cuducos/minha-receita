# Como usar

A API web tem apenas um _endpoints_ principal: `/<número do CNPJ>`. Nos exemplos a seguir, substituta `https://minhareceita.org` por `http://0.0.0.0:8000` caso esteja rodando o servidor localmente.

| Caminho da URL | Tipo de requisição | Código esperado na resposta | Conteúdo esperado na resposta |
|---|---|---|---|
| `/` | `POST` | 405 | `{"message": "Essa URL aceita apenas o método GET."}` |
| `/` | `HEAD` | 405 | `{"message": "Essa URL aceita apenas o método GET."}` |
| `/` | `GET` | 302 | _Redireciona para essa documentação._ |
| `/foobar` | `GET` | 400 | `{"message": "CNPJ foobar inválido."}` |
| `/00000000000000` | `GET` | 404 | `{"message": "CNPJ 00.000.000/0000-00 não encontrado."}`  |
| `/00.000.000/0000-00` | `GET` | 404 | `{"message": "CNPJ 00.000.000/0000-00 não encontrado."}`  |
| `/33683111000280` | `GET` | 200 | _Ver JSON de exemplo abaixo._ |
| `/33.683.111/0002-80` | `GET` | 200 | _Ver JSON de exemplo abaixo._ |

## Exemplo de requisição usando o `curl`

```console
$ curl https://minhareceita.org/33683111000280
```

## Exemplo de resposta válida

```json
{
    "cnpj": "33683111000280",
    "identificador_matriz_filial": 2,
    "descricao_identificador_matriz_filial": "FILIAL",
    "nome_fantasia": "REGIONAL BRASILIA-DF",
    "situacao_cadastral": 2,
    "descricao_situacao_cadastral": "ATIVA",
    "data_situacao_cadastral": "2004-05-22",
    "motivo_situacao_cadastral": 0,
    "descricao_motivo_situacao_cadastral": "SEM MOTIVO",
    "nome_cidade_no_exterior": "",
    "codigo_pais": null,
    "pais": null,
    "data_inicio_atividade": "1967-06-30",
    "cnae_fiscal": 6204000,
    "cnae_fiscal_descricao": "Consultoria em tecnologia da informação",
    "descricao_tipo_de_logradouro": "AVENIDA",
    "logradouro": "L2 SGAN",
    "numero": "601",
    "complemento": "MODULO G",
    "bairro": "ASA NORTE",
    "cep": "70836900",
    "uf": "DF",
    "codigo_municipio": 9701,
    "codigo_municipio_ibge": 5300108,
    "municipio": "BRASILIA",
    "ddd_telefone_1": "",
    "ddd_telefone_2": "",
    "ddd_fax": "",
    "situacao_especial": "",
    "data_situacao_especial": null,
    "opcao_pelo_simples": null,
    "data_opcao_pelo_simples": null,
    "data_exclusao_do_simples": null,
    "opcao_pelo_mei": null,
    "data_opcao_pelo_mei": null,
    "data_exclusao_do_mei": null,
    "razao_social": "SERVICO FEDERAL DE PROCESSAMENTO DE DADOS (SERPRO)",
    "codigo_natureza_juridica": 2011,
    "natureza_juridica": "Empresa Pública",
    "qualificacao_do_responsavel": 16,
    "capital_social": 1061004800,
    "codigo_porte": 5,
    "porte": "DEMAIS",
    "ente_federativo_responsavel": null,
    "descricao_porte": "",
    "qsa": [
        {
            "identificador_de_socio": 2,
            "nome_socio": "ANDRE DE CESERO",
            "cnpj_cpf_do_socio": "***220050**",
            "codigo_qualificacao_socio": 10,
            "qualificacao_socio": "Diretor",
            "data_entrada_sociedade": "2016-06-16",
            "codigo_pais": null,
            "pais": null,
            "cpf_representante_legal": "***000000**",
            "nome_representante_legal": "",
            "codigo_qualificacao_representante_legal": 0,
            "qualificacao_representante_legal": null,
            "codigo_faixa_etaria": 6,
            "faixa_etaria": "Entre 51 a 60 anos"
        },
        {
            "identificador_de_socio": 2,
            "nome_socio": "ANTONIO DE PADUA FERREIRA PASSOS",
            "cnpj_cpf_do_socio": "***595901**",
            "codigo_qualificacao_socio": 10,
            "qualificacao_socio": "Diretor",
            "data_entrada_sociedade": "2016-12-08",
            "codigo_pais": null,
            "pais": null,
            "cpf_representante_legal": "***000000**",
            "nome_representante_legal": "",
            "codigo_qualificacao_representante_legal": 0,
            "qualificacao_representante_legal": null,
            "codigo_faixa_etaria": 7,
            "faixa_etaria": "Entre 61 a 70 anos"
        },
        {
            "identificador_de_socio": 2,
            "nome_socio": "WILSON BIANCARDI COURY",
            "cnpj_cpf_do_socio": "***414127**",
            "codigo_qualificacao_socio": 10,
            "qualificacao_socio": "Diretor",
            "data_entrada_sociedade": "2019-06-18",
            "codigo_pais": null,
            "pais": null,
            "cpf_representante_legal": "***000000**",
            "nome_representante_legal": "",
            "codigo_qualificacao_representante_legal": 0,
            "qualificacao_representante_legal": null,
            "codigo_faixa_etaria": 8,
            "faixa_etaria": "Entre 71 a 80 anos"
        },
        {
            "identificador_de_socio": 2,
            "nome_socio": "GILENO GURJAO BARRETO",
            "cnpj_cpf_do_socio": "***099595**",
            "codigo_qualificacao_socio": 16,
            "qualificacao_socio": "Presidente",
            "data_entrada_sociedade": "2020-02-03",
            "codigo_pais": null,
            "pais": null,
            "cpf_representante_legal": "***000000**",
            "nome_representante_legal": "",
            "codigo_qualificacao_representante_legal": 0,
            "qualificacao_representante_legal": null,
            "codigo_faixa_etaria": 5,
            "faixa_etaria": "Entre 41 a 50 anos"
        },
        {
            "identificador_de_socio": 2,
            "nome_socio": "RICARDO CEZAR DE MOURA JUCA",
            "cnpj_cpf_do_socio": "***989951**",
            "codigo_qualificacao_socio": 10,
            "qualificacao_socio": "Diretor",
            "data_entrada_sociedade": "2020-05-12",
            "codigo_pais": null,
            "pais": null,
            "cpf_representante_legal": "***000000**",
            "nome_representante_legal": "",
            "codigo_qualificacao_representante_legal": 0,
            "qualificacao_representante_legal": null,
            "codigo_faixa_etaria": 5,
            "faixa_etaria": "Entre 41 a 50 anos"
        },
        {
            "identificador_de_socio": 2,
            "nome_socio": "ANTONINO DOS SANTOS GUERRA NETO",
            "cnpj_cpf_do_socio": "***073447**",
            "codigo_qualificacao_socio": 5,
            "qualificacao_socio": "Administrador",
            "data_entrada_sociedade": "2019-02-11",
            "codigo_pais": null,
            "pais": null,
            "cpf_representante_legal": "***000000**",
            "nome_representante_legal": "",
            "codigo_qualificacao_representante_legal": 0,
            "qualificacao_representante_legal": null,
            "codigo_faixa_etaria": 7,
            "faixa_etaria": "Entre 61 a 70 anos"
        }
    ],
    "cnaes_secundarios": [
        {
            "codigo": 6201501,
            "descricao": "Desenvolvimento de programas de computador sob encomenda"
        },
        {
            "codigo": 6202300,
            "descricao": "Desenvolvimento e licenciamento de programas de computador customizáveis"
        },
        {
            "codigo": 6203100,
            "descricao": "Desenvolvimento e licenciamento de programas de computador não-customizáveis"
        },
        {
            "codigo": 6209100,
            "descricao": "Suporte técnico, manutenção e outros serviços em tecnologia da informação"
        },
        {
            "codigo": 6311900,
            "descricao": "Tratamento de dados, provedores de serviços de aplicação e serviços de hospedagem na internet"
        }
    ]
}
```

## _Endpoints_ auxiliares

Para todos esses _endpoints_ é esperada resposta com status `200`:

| Caminho da URL | Tipo de requisição | Conteúdo esperado na resposta |
|---|---|---|
| `/updated` | `GET` | JSON contendo a data de extração dos dados pela Receita Federal. |
| `/healthz` | `GET` ou `HEAD` | Resposta sem conteúdo |
