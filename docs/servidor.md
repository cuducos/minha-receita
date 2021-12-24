# Servidor

## Banco de dados

O projeto requer um banco de dados PostgreSQL e os comandos que requerem banco de dados aceitam `--database-uri` (ou `-d`) como argumento com a URI de acesso ao PostgreSQL (o padrão é o valor da variável de ambiente `POSTGRES_URI`).

Caso deseje usar o Docker Compose do projeto para subir uma instância do banco de dados:

```console
$ docker-compose up -d postgres
```

A URI de acesso será `postgres://minhareceita:minhareceita@localhost:5432/minhareceita?sslmode=disable`.

## Download dos dados


O comando `download` baixa dados do servidor da Receita Federal, que é lento e instável. Como alternativa, pode-se apenas listar as URLs para download dos arquivos com `--urls-only` (ou `-u`). O _timeout_ padrão para o download de cada arquivo é de 15min, mas pode ser alterado com `--timeout` (ou `-t`).

Na sequência, o comando `transform` transforma os arquivos para o formato JSON (utilizado pela API web) e em um CSV unificado (utilizado para fazer a carga no banco de dados).

Para especificar onde ficam esses arquivos, os comandos aceitam como argumento:

* `--source-directory` (ou `-s`) com um diretório onde serão salvos os arquivos originais da Receita Federal
* `--out-directory` (ou `-o` com um diretório onde serão criados arquivos JSON e CSV gerados pela Minha Receita

O padrão para ambos é `data/`.

### Exemplos de uso

Sem Docker:

```console
$ minha-receita download --urls-only
$ minha-receita download --timeout 1h42m12s
$ minha-receita transform
```

Com Docker:

```console
$ docker-compose run --rm minha-receita download --source-directory /mnt/data/
$ docker-compose run --rm minha-receita transform --output-directory /mnt/data/ --source-directory /mnt/data/
```

## Carregamento do banco de dados

Primeiro é necessário criar as tabelas no banco de dados, para isso utilize o comando `create`.

Caso seja necessário limpar o banco de dados para começar um novo carregamento de dados, é possível excluir as tabelas com comando `drop`.

Para importar os dados, utilize o comando `import` — esse comando pode demorar horas, dependendo do equipamento.

### Questões de privacidade

Assim como o [`socios-brasil`](https://github.com/turicas/socios-brasil#privacidade) removemos alguns dados para evitar exposição de dados sensíveis de pessoas físicas, bem como SPAM. Ao contrário do `socios-brasil`, não temos uma opção para desativar essa camada de privacidade (mas PRs são bem-vindos).

### Exemplos de uso

Sem Docker, com a variável de ambiente `POSTGRES_URI` configurada:

```console
$ minha-receita drop  # caso necessário
$ minha-receita create
$ minha-receita import
```

Com Docker:

```console
$ docker-compose run --rm minha-receita drop  # caso necessário
$ docker-compose run --rm minha-receita create
$ docker-compose run --rm minha-receita import -o /mnt/data
```

## Iniciando a API web

A API web é uma aplicação super simples que, por padrão, ficará disponível em [`localhost:8000`](http://localhost:8000).

### Exemplos de uso

Sem Docker, com a variável de ambiente `POSTGRES_URI` configurada:

```console
$ minha-receita api
```

Com Docker:

```console
$ docker-compose up
```
