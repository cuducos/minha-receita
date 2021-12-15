# Servidor

## Download dos dados

O comando `download` faz o download dos arquivos necessários para alimentar o banco de dados. Na sequência, o comando `transform` transforma os arquivos para o formato JSON. Ambos aceitam o argumento `--directory` (ou `-d`) com um diretório onde encontrar os dados (o padrão é `data/`).

O comando `download` baixa dados do servidor da Receita Federal, que é lento e instável. Quando um download falha, nenhum arquivo é salvo (ou seja, não fica, no diretório, um arquivo pela metade; pode-se assumir que os arquivos restantes são íntegros e não precisam ser baixados novamente). Por esse motivo pode ser esperado que a barra de progresso de download recue (quando um arquivo de download falha, retiramos os bytes baixados da barra de download, pois na nova tentativa o download começa do zero).

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
$ minha-receita parse
```

Com Docker:

```console
$ docker-compose run --rm minha-receita download --directory /mnt/data/
$ docker-compose run --rm minha-receita parse --directory /mnt/data/
```

## Carregamento do banco de dados

Primeiro é necessário criar as tabelas no banco de dados, para isso utlize o comando `create`.

Caso seja necessário limpar o banco de dados para começar um novo carregamento de dados, é possível excluir as tabelas com comando `drop`.

Para importar os dados, utilize o comando `import` — esse comando pode demorar horas, dependendo do equipamento. Esse comando também aceita a opção `--directory` ou `-d` para especificar um local diferente do padrão onde encontrar os arquivos JSON gerados com o comando `transform`.

### Questões de privacidade

Assim como o [`socios-brasil`](https://github.com/turicas/socios-brasil#privacidade) removemos alguns dados para evitar exposição de dados sensíveis de pessoas físicas, bem como SPAM. Ao contrário do `socios-brasil`, não temos uma opção para desativar essa camada de privacidade (mas PRs são bem-vindos).

### Exemplos de uso

Sem Docker:

```console
$ minha-receita drop  # caso necessário
$ minha-receita create
$ minha-receita import
```

Com Docker:

```console
$ docker-compose run --rm minha-receita drop  # caso necessário
$ docker-compose run --rm minha-receita create
$ docker-compose run --rm minha-receita import -d /mnt/data/
```

## Iniciando a API web

A API web é uma aplicação super simples que, por padrão, ficará disponível em [`localhost:8000`](http://localhost:8000).

### Exemplos de uso

Sem Docker:

```console
$ minha-receita api
```

Com Docker:

```console
$ docker-compose up
```
