# Cliente

A API web tem apenas dois _endpoints_:

## `POST /`

| Caminho da URL | Tipo de requisição | Código esperado na resposta | Conteúdo esperado na resposta |
|---|---|---|---|
| `/` | `POST` | 405 | `{"message": "Essa URL aceita apenas o método GET."}` |
| `/` | `HEAD` | 405 | `{"message": "Essa URL aceita apenas o método GET."}` |
| `/` | `GET` | 400 | `{"message": "CNPJ não enviado na requisição GET."}` |
| `/foobar` | `GET` | 400 | `{"message": "CNPJ foobar inválido."}` |
| `/00000000000000` | `GET` | 204 | _Corpo vazio, significa que o CNPJ é válido mas não consta no banco de dados._ |
| `/19131243000197` | `GET` | 200 | _Ver JSON de exemplo abaixo._ |
| `/19.131.243/0001-97` | `GET` | 200 | _Ver JSON de exemplo abaixo._ |

### Exemplo de requisição usando o `curl`

```console
$ curl -i 0.0.0.0:8000/19131243000197
```

### Exemplo de resposta válida

```json
{
  "cnpj": "19131243000197",
  "identificador_matriz_filial": 1,
  "descricao_matriz_filial": "Matriz",
  "razao_social": "OPEN KNOWLEDGE BRASIL",
  "nome_fantasia": "REDE PELO CONHECIMENTO LIVRE",
  "situacao_cadastral": 2,
  "descricao_situacao_cadastral": "Ativa",
  "data_situacao_cadastral": "2013-10-03",
  "motivo_situacao_cadastral": 0,
  "nome_cidade_exterior": null,
  "codigo_natureza_juridica": 3999,
  "data_inicio_atividade": "2013-10-03",
  "cnae_fiscal": 9430800,
  "cnae_fiscal_descricao": "Atividades de associações de defesa de direitos sociais",
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
  "ddd_telefone_2": null,
  "ddd_fax": null,
  "qualificacao_do_responsavel": 10,
  "capital_social": 0.0,
  "porte": 5,
  "descricao_porte": "Demais",
  "opcao_pelo_simples": false,
  "data_opcao_pelo_simples": null,
  "data_exclusao_do_simples": null,
  "opcao_pelo_mei": false,
  "situacao_especial": null,
  "data_situacao_especial": null,
  "cnaes_secundarios": [
    {
      "codigo": 9493600,
      "descricao": "Atividades de organizações associativas ligadas à cultura e à arte"
    },
    {
      "codigo": 9499500,
      "descricao": "Atividades associativas não especificadas anteriormente"
    },
    {
      "codigo": 8599699,
      "descricao": "Outras atividades de ensino não especificadas anteriormente"
    },
    {
      "codigo": 8230001,
      "descricao": "Serviços de organização de feiras, congressos, exposições e festas"
    },
    {
      "codigo": 6204000,
      "descricao": "Consultoria em tecnologia da informação"
    }
  ],
  "qsa": [
    {
      "identificador_de_socio": 2,
      "nome_socio": "NATALIA PASSOS MAZOTTE CORTEZ",
      "cnpj_cpf_do_socio": "***059967**",
      "codigo_qualificacao_socio": 10,
      "percentual_capital_social": 0,
      "data_entrada_sociedade": "2019-02-14",
      "cpf_representante_legal": null,
      "nome_representante_legal": null,
      "codigo_qualificacao_representante_legal": null
    }
  ]
}
```

## `GET /healthz`

| Caminho da URL | Tipo de requisição | Código esperado na resposta | Conteúdo esperado na resposta |
|---|---|---|---|
| `/healthz` | `GET` | 200 | |
| `/healthz` | `HEAD` | 405 | `{"message": "Essa URL aceita apenas o método GET."}` |
| `/healthz` | `POST` | 405 | `{"message": "conteúdo inválido na requisição GET."}` |
