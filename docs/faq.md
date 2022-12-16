# Perguntas frequentes

## Como buscar por CNAE, UF, etc.?

Isso encarece computacionalmente o _Minha Receita_ (consultas mais pesadas ao banco de dados, paginação, etc.), ou seja:
* ou aumentaria os custos para manter a API no ar (e as doações não cobrem nem os custos atuais)
* ou comprometeria a disponibilidade da API.

Para evitar isso, a API não tem qualquer forma de filtro ou paginação.

No entanto, [criando o seu banco de dados localmente](servidor.md) , é possível utilizar consultas diratemente no PostgreSQL, como por exemplo:

* busca por UF com `SELECT * FROM cnpj WHERE json->>'uf' = 'PR'`
* busca por CNAE (apenas o primário, fiscal) com `SELECT * FROM cnpj WHERE json->>'cnae_fiscal' = '6204000'`
* busca por CNAE (incluindo secundários) com

```sql
SELECT *
FROM cnpj
WHERE json->>'cnae_fiscal' = '6204000'
  AND json->'cnaes_secundarios' @> '[{"codigo":6204000}]
```

Você pode ainda criar índices para essas buscas ficarem mais rápidas, como por exemplo `CREATE INDEX cnpj_uf_idx ON cnpj USING GIN ((json->'uf'))`.

## Como consigo um CSV único dos dados?

Os dados oficiais da Receita Federal já são em CSV e talvez o mais fácil seja você utilizar os CSVs originais.

Caso queira gerar um CSV a partir dos dados consolidados, com ou sem filtragem de dados, o _Minha Receita_ pode ajudar. [Criando o seu banco de dados localmente](servidor.md), é possível utilizar o PostgreSQL para tal, como por exemplo:

```sql
WITH nome_da_busca AS (
  SELECT
      id AS cnpj,
      json->>'razao_social' AS razao_social,
      json->>'nome_fantasia' AS nome_fantasia
      -- adicione os campos que desejar
  -- WHERE uf = '…' caso queira filtar, por exemplo
  FROM cnpj
)
COPY nome_da_busca
TO 'nome-do-arquivo.csv' DELIMITER ',' CSV HEADER;
```

## Não consigo baixar os dados, pode me passar?

O _Minha Receita_ mantém um [espelho dos dados em uma diretório compartilhado](https://mirror.minhareceita.org). Nele está disponível:

* Uma cópia dos públicos baixados da Receita Federal em , organizados por diretórios com a data de extração divulgada pela Receita Federal
* Executáveis do _Minha Receita_

<script data-goatcounter="https://analytics.minhareceita.org" async src="//analytics.minhareceita.org/count.js"></script>
