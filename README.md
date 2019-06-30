# Minha Receita

Sua API web para consulta de informações do CNPJ (Cadastro Nacional da Pessoa Jurídica) da Receita Federal.

## Sobre

### Histórico

Pela [Lei de Acesso à Informação](http://www.acessoainformacao.gov.br/assuntos/conheca-seu-direito/a-lei-de-acesso-a-informacao), os dados de CNPJ devem ser públicos e acessíveis por máquina. A Receita Federal oferece esses dados escondidos atrás de um CAPTCHA ou em formato pouco convencional (um _fixed-width text file_), com links lentos para download de diversos arquivos somando gigas. Isso não é acessível o suficiente.

O [Turicas](https://twitter.com/turicas) já baixou e converteu esses arquivos para CSV, e ainda oferece um _mirror_ para download mais estável desses arquivos — tudo isso no [`socios-brasil`](https://github.com/turicas/socios-brasil).

### Propósito

O código desse repositório faz esses dados ainda mais acessíveis:

1. Importando autimaticamente os dados para um banco de dados PostgreSQL utilizando a [`rows`](https://github.com/turicas/rows)
2. Adicionando dados com descrições dos CNAEs (inexistente nos arquivos da Receita Federal)
3. Fornecendo uma API web para a consulta de dados de um CNPJ

### Qual a URL para acesso?

Não tem. Disponibilizo essa aplicação para que cada um rode na sua própria infraestrutura, pois:

1. não tenho dinheiro para manter um serviço desse porte no ar
2. não tenho interesse em desenvolver um sistema para cobrar por esse serviço

## Instalação

### Requisitos

* [Docker Compose](https://docs.docker.com/compose/install/)
* Cerca de 30Gb disponíveis de espaço em disco
* _Talvez_ seja necessário um sistema UNIX (Linux ou macOS), mas não tenho certeza pois não testei em Windows.

### Download dos arquivos

Salve **quatro** arquivos no diretório `data/` desse repositório (3 da Receita Federal, 1 do IBGE):

#### Receita Federal

Primeiro faço o download dos arquivos da Receita Federal convertidos para CSV com o [`socios-brasil`](https://github.com/turicas/socios-brasil) e [disponibilizados no Google Drive do Turicas](https://drive.google.com/drive/folders/1JRJDfjm6uHqyEruJdtWT--kPwUx3APEy):

* `empresa.csv.gz`
* `socio.csv.gz`
* `cnae-secundaria.csv.gz`

#### IBGE

Depois, precisamos dos dados com a descrição dos CNAE (Classificação Nacional de Atividades Econômicas), já que nos dados da Receita Federal apenas temos o código númerico, sem a descrição. Acesse [a página da CONCLA no IBGE](https://cnae.ibge.gov.br/classificacoes/download-concla.html) e baixe a planilha _Estrutura detalhada_ do documento _CNAE 2.3 Subclasses_:

* [`CNAE_Subclasses_2_3_Estrutura_Detalhada.xlsx`](https://cnae.ibge.gov.br/images/concla/documentacao/CNAE_Subclasses_2_3_Estrutura_Detalhada.xlsx)

### Configurações

Copie o arquivo `.env.sample` como `.env` e ajuste de acordo com as suas preferências e necessidades:

* As variáveis com o prefixo `POSTGRES_` configuram os dados de acesso ao PostgreSQL que o Docker Compose sobe
* As variáveis com o prefixo `SANIC_` configuram o framework web [Sanic](https://github.com/huge-success/sanic) (se o uso for em producão é extremamente recomendável configurar `SANIC_DEBUG` para `True`)

### Alimente o banco de dados local

Existe um serviço nesse repositório chamado `feed` só para automatizar esse carregamento de dados, ele pode demorar mais de 1h, mas funciona (lembre-se de, antes, baixar os arquivos para o diretório `data/`):

```console
$ docker-compose up feed
```

#### Inicia a API web

A API web é uma aplicação super simples que pode ser inciada com:

```console
$ docker-compose up api
```

## Uso

### API web

A API web tem apenas um endpoint (`/`) que somente aceita requisições tipo `POST`:

| Caminho da URL | Tipo de requisição | Dados enviados | Código esperado na resposta | Conteúdo esperado na resposta |
|---|---|---|---|---|
| `/` | `GET` | | 405| `Error: Method GET not allowed for URL /` |
| `/` | `POST` | | 404 | `Error: CNPJ não enviado na requisição POST.` |
| `/` | `POST` | `cpf=foobar` | 404 | `Error: CNPJ não enviado na requisição POST.` 
| `/` | `POST` | `cnpj=foobar` | 404 | `Error: CNPJ foobar inválido.` |
| `/` | `POST` | `cnpj=00000000000000` | 404 | `Error: CNPJ 00000000000000 não encontrado.` |
| `/` | `POST` | `cnpj=19131243000197` | 200 | _Ver JSON de exemplo abaixo._ |
| `/` | `POST` | `cnpj=19.131.243/0001-97` | 200 | _Ver JSON de exemplo abaixo._ |

Exemplo de resposta válida:

```json
{
    "cnpj": "19131243000197",
    "identificador_matriz_filial": 1,
    "razao_social": "OPEN KNOWLEDGE BRASIL",
    "nome_fantasia": "REDE PELO CONHECIMENTO LIVRE",
    "situacao_cadastral": 2,
    "data_situacao_cadastral": 1380758400,
    "motivo_situacao_cadastral": 0,
    "nome_cidade_exterior": null,
    "codigo_natureza_juridica": 3999,
    "data_inicio_atividade": 1380758400,
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
    "ddd_telefone_2": null,
    "ddd_fax": null,
    "qualificacao_do_responsavel": 10,
    "capital_social": 0.0,
    "porte": 5,
    "opcao_pelo_simples": false,
    "data_opcao_pelo_simples": null,
    "data_exclusao_do_simples": null,
    "opcao_pelo_mei": false,
    "situacao_especial": null,
    "data_situacao_especial": null,
    "cnaes_secundarios": [
        {
            "cnae_codigo": 9493600,
            "cnae_descricao": "Atividades de organizações associativas ligadas à cultura e à arte"
        },
        {
            "cnae_codigo": 9499500,
            "cnae_descricao": "Atividades associativas não especificadas anteriormente"
        },
        {
            "cnae_codigo": 8599699,
            "cnae_descricao": "Outras atividades de ensino não especificadas anteriormente"
        },
        {
            "cnae_codigo": 8230001,
            "cnae_descricao": "Serviçoos de organização de feiras, congressos, exposições e festas"
        },
        {
            "cnae_codigo": 6204000,
            "cnae_descricao": "Consultoria em tecnologia da informação"
        }
    ],
    "qsa": [
        {
            "identificador_de_socio": 2,
            "nome_socio": "NATALIA PASSOS MAZOTTE CORTEZ",
            "cnpj_cpf_do_socio": "***059967**",
            "codigo_qualificacao_socio": 10,
            "percentual_capital_social": 0,
            "data_entrada_sociedade": 1550102400,
            "cpf_representante_legal": null,
            "nome_representante_legal": null,
            "codigo_qualificacao_representante_legal": null
        }
    ]
}
```

### Acesso ao banco de dados

É possível, também, conectar-se diretamente ao PostresSQL, pois o serviço expõe a porta `5432`. A URI, descontando alterações no `.env`, seria `postgres://minhareceita:minhareceita@localhost:5432/minhareceita`.

## Contribuindo

Escreva testes, reconstrua os containers, rode os testes e formate seu código com [Black](https://github.com/python/black):

```console
$ docker-compose build
$ docker-compose run --rm api pipenv run py.test
$ docker-compose run --rm api black . --check
$ docker-compose run --rm feed black . --check
```

## Muito obrigado

Ao [Turicas](https://twitter.com/turicas) por todo ativismo mais o trabalho de coleta, tratamento, e carinho que faz os dados serem cada vez mais acessíveis. E ao [Bruno](https://twitter.com/555112299jedi), sem o qual [nunca teríamos acesso a esses dados por menos de R$ 500 mil](https://medium.com/serenata/o-dia-que-a-receita-nos-mandou-pagar-r-500-mil-para-ter-dados-p%C3%BAblicos-8e18438f3076).
