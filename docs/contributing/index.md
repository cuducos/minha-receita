# Introdução

Os arquivos estão organizados em pacotes Go, sendo esses os quatro principais:

1. `download/` responsável pelo download dos dados (basicamente uma automação utiliZando o [`chunk`](https://github.com/cuducos/chunk))
1. `transform/` é um <abbr title="Extract, Transform, Load">ETL</abbr> que lê os arquivos baixados pelo pacote acima, processa cada um deles e salva os dados no banco de dados — entenda melhor sobre essa etapa lendo sobre os [detalhes do <abbr title="Extract, Transform, Load">ETL</abbr>](etl.md)
1. `db/` implementação dos _backends_ de banco de dados, ou seja, implementa as interfaces necessárias pelo ETL e pela API em diferentes bancos de dados (PostgreSQL é utilizado em produção)
1. `api/` a API web que serve os dados do projeto para o público

Escreva testes e rode os testes, use autoformatação e _linter_:

```console
$ gofmt ./
$ golanglint-ci run ./...
$ go test --race ./...
```

Os testes requerem uma instância de cada banco de dados implementado. Atualmente eles precisam ser configurados em `TEST_POSTGRES_URL` e `TEST_MONGODB_URL`, como no exemplo em `.env`, e podem ser [facilmente criados com o Docker Compose](docker.md).
