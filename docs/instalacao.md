# Instalação local

Existem três formas de rodar essa aplicação localmente:

* ou com a imagem Docker
* ou gerando o binário a partir do código fonte
* ou com Docker Compose — apenas para desenvolvimento (não recomendado para o banco de dados completo)

As duas últimas alternativas necessitam do código fonte. Você pode usar o Git para baixar o código do projeto:

```console
$ git clone https://github.com/cuducos/minha-receita.git
```

## Requisitos e instalação

É necessário cerca de 170Gb disponíveis de espaço em disco para armazenar os dados. Caso o banco de dados esteja em uma máquina separada, a divisão é mais ou menos 140Gb para as tabelas no banco de dados, 10Gb para os índices no banco de dados, e 20Gb para a máquina que vai fazer o download dos dados e transformá-los para carregar no banco de dados.

### Banco de dados

* O banco de dados gerado utiliza cerca de 140Gb
* Rodar o banco de dados localmente com Docker só é recomendado para desenvolvimento (não recomendado para o banco de dados completo)

#### Download e transformação dos dados

* Os arquivos da Receita federal tem menos de 10Gb
* O processo de importação utiliza uma estrutura temporária de cerca de 15Gb

### Imagem Docker

* [Docker](https://www.docker.com/)

Baixar a imagem com:

```console
$ docker pull ghcr.io/cuducos/minha-receita:main
```

### A partir do código fonte

* [Go](https://golang.org/) versão 1.23

Depois de clonar o repositório, baixe as dependências e compile a aplicação para um diretório incluído no `PATH`, por exemplo:

```console
$ go get
$ go build -o /usr/local/bin/minha-receita main.go
```

### Docker Compose

* [Docker](https://www.docker.com/)
* [Docker Compose](https://docs.docker.com/compose/install/)
* Arquivo `.env` (copie o `.env.sample` e ajuste caso necessário)

Gere as imagens dos containers com:

```console
$ docker compose build
```

## Execução e configurações

Várias configurações podem ser passadas para a CLI, e elas estão documentadas no `--help` de cada comando da aplicação.

### Exemplos

#### Imagem Docker

```console
$ docker run --rm ghcr.io/cuducos/minha-receita:main --help
$ docker run --rm ghcr.io/cuducos/minha-receita:main api --help
```

#### A partir do código fonte

```console
$ minha-receita --help
$ minha-receita api --help
```

#### Docker Compose

```console
$ docker compose run --rm minha-receita --help
$ docker compose run --rm minha-receita api --help
```

### Variáveis de ambiente

Para facilitar a manutenção, algumas variáveis de ambiente podem ser utilizadas, mas todas são opcionais:

| Variável | Descrição |
|---|---|
| `DATABASE_URL` | URI de acesso ao banco de dados |
| `PORT` | Porta na qual a API web ficará disponível |
| `NEW_RELIC_LICENSE_KEY` | Licença no New Relic para monitoramento |
| `TEST_POSTGRES_URL` | URI de acesso ao banco de dados PostgreSQL para ser utilizado nos testes |
| `TEST_MONGODB_URL` | URI de acesso ao banco de dados MongoDB para ser utilizado nos testes |
