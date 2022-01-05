# Instalação

## Requisitos

Existem duas formas de rodar essa aplicação:

* ou diretamente a partir do seu sistema operacional
* ou com Docker

_Talvez_ seja necessário um sistema UNIX (Linux ou macOS), mas não tenho certeza pois não testei em Windows.

É necessário cerca de 650Gb disponíveis de espaço em disco para armazenar os dados:
* Os arquivos da Receita federal tem cerca de 6Gb
* Os arquivos JSON gerados pela Minha Receita ocupam cerca de 400Gb (mas podem ser excluídos ao final do processo)
* O banco de dados PostgreSQL, cerca de 240Gb

### Requisitos e instalação sem Docker

* [Go](https://golang.org/) versão 1.16
* Cliente [PostgreSQL](https://www.postgresql.org/) (comando `psql` disponível no seu terminal — em sistemas Debian, `apt install postgresql-client` resolve)

Baixe as dependências e compile a aplicação para um diretório incluído no `PATH`, por exemplo:

```console
$ go get
$ go build -o /usr/local/bin minha-receita
```

### Requisitos e instalação com Docker

* [Docker](https://www.docker.com/)
* [Docker Compose](https://docs.docker.com/compose/install/)
* Arquivo `.env` (copie o `.env.sample` e ajuste caso necessário)

Gere as imagens dos containers com:

```console
$ docker-compose build
```

## Configurações

Várias configurações podem ser passadas para a CLI, e elas estão documentadas no `--help` de cada comando da aplicação.

### Exemplo

Sem Docker, por exemplo:

```console
$ minha-receita --help
$ minha-receita api --help
```

Com Docker, por exemplo:

```console
$ docker-compose run --rm minha-receita --help
$ docker-compose run --rm minha-receita api --help
```

### Variáveis de ambiente

Para facilitar a manutenção, algumas variáveis de ambiente podem ser utilizadas, mas todas são opcionais:

| Variável | Descrição |
|---|---|
| `POSTGRES_URI` | URI de acesso ao banco de dados PostgreSQL |
| `PORT` | Porta na qual a API web ficará disponível |
| `NEW_RELIC_LICENSE_KEY` | Licença no New Relic para monitoramento |
| `TEST_POSTGRES_URI` | URI de acesso ao banco de dados PostgreSQL para ser utilizado nos testes |
