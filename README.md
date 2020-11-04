# Minha Receita [![Tests](https://github.com/cuducos/minha-receita/workflows/API/badge.svg)](https://github.com/cuducos/minha-receita/actions)

API web para consulta de informações do CNPJ (Cadastro Nacional da Pessoa Jurídica) da Receita Federal.

1. [Sobre](#sobre)
    1. [Histórico](#historico)
    1. [Propósito](#proposito)
    1. [Qual a URL para acesso?](#qual-a-url-para-acesso)
1. [Instalação](#instalacao)
    1. [Requisitos](#requisitos)
    1. [Configurações](#configuracoes)
1. [Uso](#uso)
    1. [Download dos dados](#download-dos-dados)
    1. [Carregamento do banco de dados](#carregamento-o-banco-de-dados)
    1. [API web](#api-web)
1. [Contribuindo](#contribuindo)
1. [Muito obrigado](#muito-obrigado)

## Sobre

### Histórico

Pela [Lei de Acesso à Informação](http://www.acessoainformacao.gov.br/assuntos/conheca-seu-direito/a-lei-de-acesso-a-informacao), os dados de CNPJ devem ser públicos e acessíveis por máquina. A Receita Federal oferece esses dados escondidos atrás de um CAPTCHA ou em formato pouco convencional (um _fixed-width text file_), com links lentos e instáveis para download arquivos somando gigas. Isso não é acessível o suficiente.

O [Turicas](https://twitter.com/turicas) já baixou e converteu esses arquivos para CSV, e ainda oferece um _mirror_ para download mais estável desses arquivos — tudo isso no [`socios-brasil`](https://github.com/turicas/socios-brasil) e disponibilizado no [Brasil.IO](https://brasil.io/).

### Propósito

O código desse repositório faz esses dados ainda mais acessíveis:

1. Importando autimaticamente os dados para um banco de dados PostgreSQL
2. Adicionando dados com descrições dos CNAEs (inexistente nos arquivos da Receita Federal)
3. Fornecendo uma API web para a consulta de dados de um CNPJ

### Qual a URL para acesso?

Não tem.

Disponibilizo essa aplicação para que cada um rode na sua própria infraestrutura, pois:

1. não tenho dinheiro para manter um serviço desse porte no ar
2. não tenho interesse em desenvolver um sistema para cobrar por esse serviço

## Instalação

### Requisitos

Existem duas formas de rodar essa aplicação:

* ou diretamente a partir do seu sistema operacional
* ou com Docker

_Talvez_ seja necessário um sistema UNIX (Linux ou macOS), mas não tenho certeza pois não testei em Windows.

É necessário cerca de 30Gb disponíveis de espaço em disco para armazenar os dados.

#### Requisitos e instalação sem Docker

* [Go](https://golang.org/) versão 1.15
* Cliente [PostgreSQL](https://www.postgresql.org/) (comando `psql` disponível no seu terminal — em sistemas Debian, `apt install postgresql-client` resolve)
* Variável de ambiente`POSTGRES_URI` com crdenciais de acesso a um banco de dados PostgreSQL (como no exemplo `.env.sample`)

Baixe as dependências e compile a aplicação para um diretório incluído no set `PATH`:

```console
$ go get
$ go build -o /usr/local/bin minha-receita
```

#### Requisitos e instalação com Docker

* [Docker](https://www.docker.com/)
* [Docker Compose](https://docs.docker.com/compose/install/)

Gere as imagens dos containers com:

```console
$ docker-compose build
```

### Configurações

Todas as configurações são passadas via variáveis de ambiente e estão documentadas no `--help` da aplicação.

#### Exemplo

Sem Docker:

```console
$ minha-receita --help
$ minha-receita api --help
```

Com Docker:

```console
$ docker-compose run --rm minha-receita --help
$ docker-compose run --rm minha-receita api --help
```

## Uso

### Download dos dados

Precisamos de **quatro** arquivos no diretório `data/` desse repositório (3 da Receita Federal, 1 do IBGE). O comando `download` mostra as instruções e as URLs (no futuro ele fará o download em si).

#### Exemplo

Sem Docker:

```console
$ minha-receita download
```

Com Docker:

```console
$ docker-compose run --rm minha-receita download --directory /mnt/data/
```

### Carregamento do banco de dados

Primeiro é necessário criar as tabelas no banco de dados, para isso utlize o comando `create`.

Caso seja necessário limpar o banco de dados para começar um novo carregamento de dados, é possível excluir as tabelas com comando `drop`.

Para importar os dados, utilize o comando `import` — esse comando pode demorar mais de 1h, depdendendo do equipamento. Esse comando aceita o argumento `--directory` (ou `-d`) com um diretório onde encontrar os dados (o padrão é `./data/`).

#### Exemplos de uso

Sem Docker:

```console
$ minha-receita drop  # caso necessário
$ minha-receita create
$ minha-receita import
```

Com Docker:

```console
$ docker-compose run --rm minha-receita drop  # caso necessário
$ docker-compose run --rm minha-receita create
$ docker-compose run --rm minha-receita import -d /mnt/data/
```

### API web

A API web é uma aplicação super simples que, por padrão, ficará disponível em [`localhost:8000`](http://localhost:8000).

#### Exemplos de uso

Sem Docker:

```console
$ minha-receita api
```

Com Docker:

```console
$ docker-compose up
```

A API web tem apenas um endpoint (`/`) que somente aceita requisições tipo `POST`:

| Caminho da URL | Tipo de requisição | Dados enviados | Código esperado na resposta | Conteúdo esperado na resposta |
|---|---|---|---|---|
| `/` | `GET` | | 405| `{"message": "Essa URL aceita apenas o método POST."}` |
| `/` | `POST` | | 400 | `{"message": "conteúdo inválido na requisição POST."}` |
| `/` | `POST` | `cpf=foobar` | 400 | `{"message": "CNPJ não enviado na requisição POST."}` |
| `/` | `POST` | `cnpj=foobar` | 400 | `{"message": "CNPJ foobar inválido."}` |
| `/` | `POST` | `cnpj=00000000000000` | 404 | `{"message": "CNPJ 00.000.000/0001-91 não encontrado."}` |
| `/` | `POST` | `cnpj=19131243000197` | 200 | _Ver JSON de exemplo abaixo._ |
| `/` | `POST` | `cnpj=19.131.243/0001-97` | 200 | _Ver JSON de exemplo abaixo._ |

##### Exemplo de requisição usando o `curl`

```console
$ curl -i -X POST -d cnpj=19131243000197 0.0.0.0:8000
```

##### Exemplo de resposta válida

```json
{
  "cnpj": "19131243000197",
  "identificador_matriz_filial": 1,
  "razao_social": "OPEN KNOWLEDGE BRASIL",
  "nome_fantasia": "REDE PELO CONHECIMENTO LIVRE",
  "situacao_cadastral": 2,
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

## Contribuindo

Escreva testes, rode os testes e reconstrua os containers para saber se está tudo certo:

```console
$ go test ./...
$ docker-compose build
```

## Muito obrigado

Ao [Turicas](https://twitter.com/turicas) por todo ativismo mais o trabalho de coleta, tratamento, e carinho que faz os dados serem cada vez mais acessíveis. E ao [Bruno](https://twitter.com/555112299jedi), sem o qual [nunca teríamos acesso a esses dados por menos de R$ 500 mil](https://medium.com/serenata/o-dia-que-a-receita-nos-mandou-pagar-r-500-mil-para-ter-dados-p%C3%BAblicos-8e18438f3076).
