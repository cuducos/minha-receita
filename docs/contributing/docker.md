# Utilizando o Docker

## Apenas para o banco de dados

Caso queira utilizar o Docker apenas para subir o banco de dados (recomendado), utilize:

```console
$ docker compose up -d postgres
$ docker compose up -d mongo
```

Existe também opções de bancos de dados para teste, que não persistem dados:

```console
$ docker compose up -d postgres_test mongo_test
```

Para visualizar as queries efetuadas:

```console
$ docker compose logs postgres_test mongo_test
```

As configurações padrão desses bancos são:

| Serviço | Ambiente | Variável de ambiente | Valor |
|---|---|---|---|
| `postgres` | Desenvolvimento | `DATABASE_URL` | `postgres://minhareceita:minhareceita@localhost:5432/minhareceita?sslmode=disable` |
| `mongo` | Desenvolvimento | `DATABASE_URL` | `mongodb://minhareceita:minhareceita@localhost:27017/minhareceita?authSource=admin` |
| `postgres_test` | Testes | `TEST_POSTGRES_URL` | `postgres://minhareceita:minhareceita@localhost:5555/minhareceita?sslmode=disable` |
| `mongo_test` | Testes | `TEST_MONGODB_URL` | `mongodb://minhareceita:minhareceita@localhost:27117/minhareceita?authSource=admin` |

## Rodando o projeto todo com Docker

!!! warning "Aviso"
    O ETL não costuem funcionar com Docker. Mas, depois de carregar os dados, rodar o banco de dados e a API com Docker normalmente funciona.

Se for utilizar Docker para rodar o projeto todo, copie o arquivo `.env.sample` como `.env` — e ajuste, se necessário.

O banco de dados de sua escolha (padrão, que persiste dados; ou de testes, que não persiste dados) tem que ser [iniciado isoladamente](#apenas-para-o-banco-de-dados).
