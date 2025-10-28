# Dados e desenvolvimento

Para utilizar o Minha Receita é preciso seguir os passos para [criar o próprio servidor](../servidor.md), mas como o processo todo de [ETL](etl.md) (o comando `transform`) demora demais, caso queira testar manualmente com uma **amostra** dos dados, utilize o comando `sample` para gera arquivos limitados a 10 mil linhas (assim o processo todo roda em cerca de 1 minuto, por exemplo). Após [fazer o download dos dados](https://docs.minhareceita.org/servidor/#download-dos-dados):

```console
$ ./minha-receita sample
$ ./minha-receita transform -d data/sample
```

Explore mais opções com `--help`.

Inconsistências podem acontecer no banco de dados de testes, e `./minha-receita drop -u ` usando `$TEST_POSTGRES_URL` e `$TEST_MONGODB_URL`   é uma boa forma de evitar isso.
