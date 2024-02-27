# Criando seu próprio servidor

## Banco de dados

O projeto requer um banco de dados PostgreSQL e os comandos que requerem banco de dados aceitam `--database-uri` (ou `-u`) como argumento com a URI de acesso ao PostgreSQL (o padrão é o valor da variável de ambiente `DATABASE_URL`).

Caso deseje usar o Docker Compose do projeto para subir uma instância do banco de dados:

```console
$ docker-compose up -d postgres
```

A URI de acesso será `postgres://minhareceita:minhareceita@localhost:5432/minhareceita?sslmode=disable`.

## Download dos dados

O comando `download` baixa dados da Receita Federal, mais um arquivo do Tesouro Nacional com o código dos municípios do IBGE.

O servidor da Receita Federal é lento e instável, então todo os arquivos são [baixados em pequenas fatias](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Range).

O comando aceita um opção `--directory` (ou `-d`) com um diretório onde serão salvos os arquivos originais da Receita Federal. O padrão é `data/`.

Caso o download falhe, é recomendado variar as configurações explicadas no `--help`, por exemplo:

* número de downloads paralelos com o `--parallel` (ou `-p`)
* números de tentativas de download de cada fatia de cada arquivo com `--retries` (ou `-r`)
* tempo limite para cada fatia com `--timeout` (ou `-t`)
* rodar o comando de download sucessivas vezes com a opção `--skip` (ou `-x`) para baixar apenas os arquivos que estão faltando

Em último caso, é possível listar as URLs para download dos arquivos com comando `urls`; e, então, tentar fazer o download de outra forma (manualmente, com alguma ferramenta que permite recomeçar downloads interrompidos, etc.).

### Espelho dos dados

O _Minha Receita_ mantém um [espelho dos dados em uma diretório compartilhado](https://mirror.minhareceita.org). Você pode fazer o download dos arquivos de lá (ao invés de utilisar o servidor oficial) com a opção `--mirror YYYY-MM-DD` substituindo a data por alguma das disponíveis no espelho.

### Exemplos de uso

Sem Docker:

```console
$ minha-receita download --urls-only
$ minha-receita download --timeout 1h42m12s
$ minha-receita download --mirror 2022-12-17
```

Com Docker:

```console
$ docker-compose run --rm minha-receita download --directory /mnt/data/
```

## Verificação dos downloads

O servidor da Receita Federal, além de lento e instável, não oferece uma opção de [soma de verificação](https://pt.wikipedia.org/wiki/Soma_de_verifica%C3%A7%C3%A3o). Com isso, pode acontecer de os arquivos baixados estarem corrompidos. O comando `check` verifica a integridade dos arquivos `.zip` baixados. A opção `--delete` exclui os arquivos que falharem na verificação.

## Tratamento dos dados

O comando `transform` transforma os arquivos para o formato JSON, consolidando as informações de todos os arquivos CSV. Esse JSON é armazenado diretamente no banco de dados. Para tanto, é preciso criar a tabela no banco de dados com o comando `create` (o comando `drop` pode ser utilizado para excluir essa mesma tabela).

Para especificar onde ficam os arquivos originais da Receita Federal e do Tesouro Nacional, o comando aceita como argumento `--directory` (ou `-d`), sendo o padrão `data/`.

### Exemplos de uso

Sem Docker, com a variável de ambiente `DATABASE_URL` configurada:

```console
$ minha-receita drop  # caso necessário
$ minha-receita create
$ minha-receita transform
```

Com Docker:

```console
$ docker-compose run --rm minha-receita drop  # caso necessário
$ docker-compose run --rm minha-receita create
$ docker-compose run --rm minha-receita transform -d /mnt/data/
```

### Questões de privacidade

Assim como o [`socios-brasil`](https://github.com/turicas/socios-brasil#privacidade) removemos alguns dados para evitar exposição de dados sensíveis de pessoas físicas, bem como SPAM. A opção `--no-privacy` do comando `transform` remove essa precaução de privacidade.


## Iniciando a API web

A API web é uma aplicação super simples que, por padrão, ficará disponível em [`localhost:8000`](http://localhost:8000).

### Exemplos de uso

Sem Docker, com a variável de ambiente `DATABASE_URL` configurada:

```console
$ minha-receita api
```

Com Docker:

```console
$ docker-compose up
```
