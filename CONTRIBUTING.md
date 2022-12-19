# Contribuindo com a Minha Receita

Escreva testes e rode os testes, use autoformatação e _linter_:

```console
$ gofmt ./
$ staticcheck ./...
$ go test ./...
```

Os testes requerem um banco de dados de teste, com acesso configurado em `TEST_DATABASE_URL` como no exemplo em `.env`.

## Docker

### Apenas para o banco de dados

Caso queira utilizar o Docker apenas para subir o banco de dados, utilize:

```console
$ docker-compose up -d postgres
```

Existe também um banco de dados para teste, que não persiste dados e que loga todas as queries:

```console
$ docker-compose up -d postgres_test
```

Para visualizar as queries efetuadas:

```console
$ docker-compose logs postgres_test
```

As configurações padrão desses bancos são:

| Serviço | Ambiente | Variável de ambiente | Valor |
|---|---|---|---|
| `postgres` | Desenvolvimento | `DATABASE_URL` | `postgres://minhareceita:minhareceita@localhost:5432/minhareceita?sslmode=disable` |
| `postgres_test` | Testes | `TEST_DATABASE_URL` | `postgres://minhareceita:minhareceita@localhost:5555/minhareceita?sslmode=disable` |

### Rodando o projeto todo com Docker

Se for utilizar Docker para rodar o projeto todo,  copie o arquivo `.env.sample` como `.env` — e ajuste, se necessário.

O banco de dados de sua escolha (padrão, que persiste dados; ou de testes, que não persiste dados) tem que ser [iniciado isoladamente](#apenas-para-o-banco-de-dados).

## Arquitetura: número do CNPJ e estrutura do pacote `transform`

Todos os dados manipulados por esse pacote vem da [Receita Federal](https://dados.gov.br/dados/conjuntos-dados/cadastro-nacional-da-pessoa-juridica-cnpj).

### Contexto

Um número de CNPJ tem 3 partes, e isso é importante pois influencia a forma que a Receita Federal disponibiliza os dados:

* base
* ordem
* dígitos verificadores

Por exemplo, em `19.131.243/0001-97` o número base é `19.131.243`, a ordem é `0001` e `97` são os dígitos verificadores.

Uma mesma pessoa jurídica tem sempre a mesma base, e só varia a ordem (nas filiais dessa mesma pessoa jurídica, por exemplo), e os dígitos verificadores.

### Dados

O grosso dos dados está nos arquivos CSV de estabelecimentos que tem `Estabelecimentos*` como prefixo, e as linhas desses arquivos tem um número de CNPJ completo como chave.

#### Dados que tem o CNPJ base (apenas 8 primeiros dígitos do número de CNPJ) como chave

* Arquivos com o prefixo `Empresas*` tem o básico dos dados, como razão social, natureza jurídica e porte.
* Arquivos com o prefixo `Socios*` tem informações sobre o quadro societário de cada pessoa jurídica.
* Arquivo `Simples.zip` tem informações sobre adesão das pessoas jurídicas ao Simples e MEI.

#### Dados com outras chaves

Na leitura desses arquivos existem campos que contém um código numérico, mas sem descrição do significado (por exemplo, temos o código 9701 para o município de Brasília). Esses arquivos são chamados de tabelas de _look up_:

* Arquivo `Cnaes.zip` com descrição dos CNAEs
* Arquivo `Motivos.zip` com descrição dos motivos cadastrais
* Arquivo `Municipios.zip` com o nome dos municípios
* Arquivo `Paises.zip` com o nome dos países
* Arquivo `Naturezas.zip` com o nome da natureza jurídica
* Arquivo `Qualificacoes.zip` com a descrição da qualificação de cada pessoa do quadro societário
* [Arquivo do Tesouro Nacional com os códigos dos municípios do IBGE](https://www.tesourotransparente.gov.br/ckan/dataset/lista-de-municipios-do-siafi/resource/eebb3bc6-9eea-4496-8bcf-304f33155282)

### Estratégia de carregamento dos dados no PostgreSQL

A etapa de transformação dos dados, começa criando armazenamentos de chave e valor, com acesso rápido, para completar os dados dos CSVs principais, `Estabelecimentos*`. Isso é feito em memória para os dados que tem outras chaves, e em disco para os dados que tem como chave a base do CNPJ.

A partir daí, cada linha dos `Estabelecimentos*` é lida, enriquecida com esses pares de chave e valor armazenados anteriormente, e então enviada para o banco de dados.

Resumindo:

1. Armazena pares de chave e valor em memória para os dados de: `Cnaes.zip`, `Motivos.zip`, `Municipios.zip`, `Paises.zip`, `Naturezas.zip`, `Qualificacoes.zip` e códigos dos municípios do IBGE
1. Armazena pares de chave e valor em disco para os dados de:
    1. `Empresas*` enriquecidas com pares de chave e valor de `Cnaes.zip`, `Motivos.zip`, `Municipios.zip`, `Paises.zip`, `Naturezas.zip`, `Qualificacoes.zip` e códigos dos municípios do IBGE
    1. `Socios*` enriquecidos com pares de chave e valor de `Qualificacoes.zip`
    1. `Simples.zip` e enriquecer as linhas do banco de dados com essas informações
1. Lê os arquivos `Estabelecimentos*`
1. “Enriquece” cada linha deles com os pares de chave e valor
1. Persiste essa informação no banco de dados

## Amostra dos arquivos para testes

Como o processo todo de ETL (o comando `transform`) demora demais, caso queira testar manualmente com uma **amostra** dos dados, utilize o comando `sample` para gera arquivos limitados a 10 mil linhas (assim o processo todo roda em cerca de 1 minuto, por exemplo). Após fazer o download dos dados:

```console
$ ./minha-receita sample
$ ./minha-receita transform -d data/sample
```

Explore mais opções com `--help`.

Inconsistências podem acontecer no banco de dados de testes, e `./minha-receita drop -u $TEST_DATABASE_URL` é uma boa forma de evitar isso.

## Documentação

Utilizamos o [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/):

```console
$ docker pull squidfunk/mkdocs-material
$ docker run --rm -v $(pwd):/docs squidfunk/mkdocs-material build
```

A documentação vai ser gerada em `site/index.html`. Para servir enquanto desenvolve:

```console
$ docker run -p 8000:8000 --rm -v $(pwd):/docs squidfunk/mkdocs-material serve --dev-addr 0.0.0.0:8000
```
