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

Todos os dados manipulados por esse pacote vem da [Receita Federal](https://www.gov.br/receitafederal/pt-br/assuntos/orientacao-tributaria/cadastros/consultas/dados-publicos-cnpj).

### Contexto

Um número de CNPJ tem 3 partes, e isso é importante pois influencia a forma que a Receita Federal disponibiliza os dados:

* base
* ordem
* dígitos verificadores

Por exemplo, em `19.131.243/0001-97` o número base é `19.131.243`, a ordem é `0001` e `97` são os dígitos verificadores.

Uma mesma pessoa jurídica tem sempre a mesma base, e só varia a ordem (nas filiais dessa mesma pessoa jurídica, por exemplo), e os dígitos verificadores.

### Dados

O grosso dos dados está nos arquivos CSV de estabelecimentos que tem `ESTABELE` como sufixo.

#### Dados que tem o CNPJ base (8 primeiros dígitos do número de CNPJ) como chave

* Arquivos CSV com o sufixo `EMPRECSV` tem o básico dos dados, como razão social, natureza jurídica e porte.
* Arquivos CSV com o sufixo `SOCIOCSV` tem informações sobre o quadro societário de cada pessoa jurídica.
* Arquivos CSV com o sufixo `SIMPLES` tem informações sobre adesão das pessoas jurídicas ao Simples.

#### Dados com outras chaves

Na leitura desses arquivos existem campos que contém um código numérico, mas sem descrição do significado (por exemplo, temos o código 9701 para o município de Brasília). Esses arquivos são chamados de tabelas de _look up_:

* Arquivos CSV com o sufixo `CNAECSV` com descrição dos CNAEs
* Arquivos CSV com o sufixo `MOTICSV` com descrição dos motivos cadastrais
* Arquivos CSV com o sufixo `MUNICCSV` com o nome dos municípios
* Arquivos CSV com o sufixo `PAISCSV` com o nome dos países
* Arquivos CSV com o sufixo `NATJUCSV` com o nome da natureza jurídica
* Arquivos CSV com o sufixo `QUALSCSV` com a descrição da qualificação de cada pessoa do quadro societário
* [Arquivo do Tesouro Nacional com os códigos dos municípios do IBGE](https://www.tesourotransparente.gov.br/ckan/dataset/lista-de-municipios-do-siafi/resource/eebb3bc6-9eea-4496-8bcf-304f33155282)

### Estratégia

A etapa de transformação dos dados cria uma linha no banco de dados para cada CNPJ listado em `ESTABELE`, e depois “enriquece” essa linha com os CSVs auxiliares:

1. Ler os arquivos CSV com o sufixo `ESTABELE` e criar um registro por CNPJ completo
    1. Incorporar nessa leitura as informações das tabelas de _look up_ `CNAECSV`, `MOTICSV`, `MUNICCSV`, `PAISCSV` e códigos dos municípios do IBGE
1. Ler os arquivos CSV com sufixo `EMPRECSV` e enriquecer as linhas do banco de dados com essas informações
    1. Incorporar nessa leitura as informações da tabela de _look up_ `NATJUCSV`
1. Ler os arquivos CSV com sufixo `SOCIOCSV` e enriquecer as linhas do banco de dados com essas informações
    1. Incorporar nessa leitura as informações da tabela de _look up_ `QUALSCSV`
1. Ler os arquivos CSV com sufixo `SIMPLES` e enriquecer as linhas do banco de dados com essas informações

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
