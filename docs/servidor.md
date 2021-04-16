# Servidor

## Download dos dados

O comando `download` faz o download dos arquivos necessários para alimentar o banco de dados. Na sequência, o comando `parse` transforma os arquivos para o formato CSV. Ambos aceitam o argumento `--directory` (ou `-d`) com um diretório onde encontrar os dados (o padrão é `data/`).

O comando `download` baixa dados do servidor da Receita Federal, que é lento e instável.

### Exemplos de uso

Sem Docker:

```console
$ minha-receita download
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

Para importar os dados, utilize o comando `import` — esse comando pode demorar horas, depdendendo do equipamento. Esse comando também aceita a opção `--directory` ou `-d` para especificar um local diferente do padrão onde encontrar os arquivos.

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
