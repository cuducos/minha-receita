# Instalação

## Requisitos

Existem duas formas de rodar essa aplicação:

* ou diretamente a partir do seu sistema operacional
* ou com Docker

_Talvez_ seja necessário um sistema UNIX (Linux ou macOS), mas não tenho certeza pois não testei em Windows.

É necessário cerca de 30Gb disponíveis de espaço em disco para armazenar os dados.

### Requisitos e instalação sem Docker

* [Go](https://golang.org/) versão 1.15
* Cliente [PostgreSQL](https://www.postgresql.org/) (comando `psql` disponível no seu terminal — em sistemas Debian, `apt install postgresql-client` resolve)
* Variável de ambiente`POSTGRES_URI` com credenciais de acesso a um banco de dados PostgreSQL (como no exemplo `.env.sample`)

Baixe as dependências e compile a aplicação para um diretório incluído no `PATH`, por exemplo:

```console
$ go get
$ go build -o /usr/local/bin minha-receita
```

### Requisitos e instalação com Docker

* [Docker](https://www.docker.com/)
* [Docker Compose](https://docs.docker.com/compose/install/)

Gere as imagens dos containers com:

```console
$ docker-compose build
```

## Configurações

Gere o arquivo com as variáveis de ambiente baseado no arquivo de exemplo:

```console
$ cp .env.sample .env
```

Todas as configurações podem ser sobrescritas via variáveis de ambiente e estão documentadas no `--help` da aplicação.

### Exemplo

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
