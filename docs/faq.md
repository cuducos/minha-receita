# Perguntas frequentes

## Como buscar por CNAE, UF, etc.?

Isso encarece computacionalmente o _Minha Receita_ (consultas mais pesadas ao banco de dados, paginação, etc.), ou seja:

* ou aumentaria os custos para manter a API no ar (e as doações não cobrem nem os custos atuais)
* ou comprometeria a disponibilidade da API.

Para evitar isso, a API não tem qualquer forma de filtro ou paginação.

No entanto, [criando o seu banco de dados localmente](servidor.md), é possível utilizar consultas diratemente no PostgreSQL, como por exemplo:

* busca por UF com `SELECT * FROM cnpj WHERE json->>'uf' = 'PR'`
* busca por CNAE (apenas o primário, fiscal) com `SELECT * FROM cnpj WHERE json->>'cnae_fiscal' = '6204000'`
* busca por CNAE (incluindo secundários) com

```sql
SELECT *
FROM cnpj
WHERE json->>'cnae_fiscal' = '6204000'
   OR json->'cnaes_secundarios' @> '[{"codigo":6204000}]
```

Você pode ainda criar índices para essas buscas ficarem mais rápidas com o comando `extra-indexes`. O comando aceita um ou mais índices, e o nome dos índices é composto pelas chaves do JSON separadas por `.`. Por exemplo, para criar um índice para a UF e para os códigos dos CNAEs secundários:

```console
$ minha-receita extra-indexes uf cnaes_secundarios.codigo
```


Para referência, no PostgreSQL:

* um índice criado apenas com o código do CNAE fiscal ocupou certa de 2Gb em disco
* um índice composto com UF e código do município ocupou cerca de 1,5Gb em disco.

## Como consigo um CSV único dos dados?

Os dados oficiais da Receita Federal já são em CSV e talvez o mais fácil seja você utilizar os CSVs originais.

Caso queira gerar um CSV a partir dos dados consolidados, com ou sem filtragem de dados, o _Minha Receita_ pode ajudar. [Criando o seu banco de dados localmente](servidor.md), é possível utilizar o PostgreSQL para tal, como por exemplo:

```sql
COPY (
  SELECT
      id AS cnpj,
      json->>'razao_social' AS razao_social,
      json->>'nome_fantasia' AS nome_fantasia
      -- adicione os campos que desejar
  -- WHERE uf = '…' caso queira filtar, por exemplo
  FROM cnpj
)
TO 'nome-do-arquivo.csv' DELIMITER ',' CSV HEADER;
```

## Com qual periodicidade a API é atualizada?

A atualização é manual e normalmente ocorre alguns dias depois de a Receita Federal liberar uma nova versão dos dados — salvo quando a Receita Federal divulga dados absurdos, como empresas abertas no ano [202](https://twitter.com/cuducos/status/1646684441979281410) ou [4100](https://twitter.com/cuducos/status/1479078346248097793).

Para saber a data de divulgação pela Receita Federal dos dados disponíveis no _Minha Receita_, [consulte a própria API](como-usar.md#endpoints-auxiliares).

## Não consigo baixar os dados, pode me passar?

O _Minha Receita_ mantém um [espelho dos dados em uma diretório compartilhado](https://mirror.minhareceita.org). Nele está disponível:

* Uma cópia dos dados públicos baixados da Receita Federal, organizados por diretórios com a data de extração divulgada pela Receita Federal
* Executáveis do _Minha Receita_
* Uma versão dos links em JSON enviando o cabeçalho `Accept: application/json` ao servidor
