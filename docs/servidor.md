# Criando seu próprio servidor

## Banco de dados

O projeto requer um banco de dados PostgreSQL e os comandos que requerem banco de dados aceitam `--database-uri` (ou `-u`) como argumento com a URI de acesso ao PostgreSQL (o padrão é o valor da variável de ambiente `POSTGRES_URI`).

Caso deseje usar o Docker Compose do projeto para subir uma instância do banco de dados:

```console
$ docker-compose up -d postgres
```

A URI de acesso será `postgres://minhareceita:minhareceita@localhost:5432/minhareceita?sslmode=disable`.

## Download dos dados

O comando `download` baixa dados da Receita Federal, mais um arquivo do Tesouro Nacional com o código dos municípios do IBGE.

O servidor da Receita Federal é lento e instável, e quando um download falha, o arquivo não é salvo (ou seja, não fica, no diretório, um arquivo pela metade; pode-se assumir que os arquivos restantes são íntegros e não precisam ser baixados novamente). Por esse motivo pode ser esperado que a barra de progresso de download recue (quando um arquivo de download falha, retiramos os bytes baixados da barra de download, pois na nova tentativa o download começa do zero).

O comando aceita um opção `--directory` (ou `-d`) com um diretório onde serão salvos os arquivos originais da Receita Federal. O padrão é `data/`.

Caso o download falhe, é recomendado variar as configurações explicadas no `--help`, por exemplo:

* diminuir o número de downloads paralelos com o `--parallel` (ou `-p`)
* aumentar o números de tentativas de download de um mesmo arquivo com `--retries` (ou `-r`)
* aumentar o tempo de `--timeout` (ou `-t`)
* rodar o comando de download sucessivas vezes com a opção `--skip` (ou `-x`) para baixar apenas os arquivos que estão faltando
* por fim, pode-se apenas listar as URLs para download dos arquivos com `--urls-only` (ou `-u`) e tentar fazer o download de outra forma (manualmente, com alguma ferramenta que permite recomeçar downloads interrompidos, etc.)

### Exemplos de uso

Sem Docker:

```console
$ minha-receita download --urls-only
$ minha-receita download --timeout 1h42m12s
```

Com Docker:

```console
$ docker-compose run --rm minha-receita download --directory /mnt/data/
```

## Verificação dos downloads

O comando `check` verifica a integridade dos arquivos `.zip` baixados. A opção `--delete` exclui os arquivos que falharem na verificação.

É possível rodar o comando `check` e o comando `download` de forma cíclica e automática até que todos os arquivos estejam baixados e íntegros: isso é feito utilizando a opção no `--insist` no comando `download`.

## Tratamento dos dados

O comando `transform` transforma os arquivos para o formato JSON, consolidando as informações de todos os arquivos CSV. Esse JSON é armazenado diretamente no banco de dados. Para tanto, é preciso criar a tabela no banco de dados com o comando `create` (o comando `drop` pode ser utilizado para excluir essa mesma tabela).

Para especificar onde ficam os arquivos originais da Receita Federal e do Tesouro Nacional, o comando aceita como argumento `--directory` (ou `-d`), sendo o padrão `data/`.

### Exemplos de uso

Sem Docker, com a variável de ambiente `POSTGRES_URI` configurada:

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

Sem Docker, com a variável de ambiente `POSTGRES_URI` configurada:

```console
$ minha-receita api
```

Com Docker:

```console
$ docker-compose up
```
