# Como usar

A API web tem apenas um _endpoints_ principal: `/<número do CNPJ>`. Nos exemplos a seguir, substituta `https://minhareceita.org` por `http://0.0.0.0:8000` caso esteja rodando o servidor localmente.

| Caminho da URL | Tipo de requisição | Código esperado na resposta | Conteúdo esperado na resposta |
|---|---|---|---|
| `/` | `POST` | 405 | `{"message": "Essa URL aceita apenas o método GET."}` |
| `/` | `HEAD` | 405 | `{"message": "Essa URL aceita apenas o método GET."}` |
| `/` | `GET` | 302 | _Redireciona para essa documentação._ |
| `/foobar` | `GET` | 400 | `{"message": "CNPJ foobar inválido."}` |
| `/00000000000000` | `GET` | 404 | `{"message": "CNPJ 00.000.000/0000-00 não encontrado."}`  |
| `/00.000.000/0000-00` | `GET` | 404 | `{"message": "CNPJ 00.000.000/0000-00 não encontrado."}`  |
| `/33683111000280` | `GET` | 200 | _Ver JSON de exemplo abaixo._ |
| `/33.683.111/0002-80` | `GET` | 200 | _Ver JSON de exemplo abaixo._ |
| `/?uf=SP` | `GET` | 200 | _Busca paginada, ver detalhes abaixo._ |

## Exemplo de requisição usando o `curl`

```console
$ curl https://minhareceita.org/33683111000280
```

## Exemplo de resposta válida

```json
{
    "cnpj": "33683111000280",
    "identificador_matriz_filial": 2,
    "descricao_identificador_matriz_filial": "FILIAL",
    "nome_fantasia": "REGIONAL BRASILIA-DF",
    "situacao_cadastral": 2,
    "descricao_situacao_cadastral": "ATIVA",
    "data_situacao_cadastral": "2004-05-22",
    "motivo_situacao_cadastral": 0,
    "descricao_motivo_situacao_cadastral": "SEM MOTIVO",
    "nome_cidade_no_exterior": "",
    "codigo_pais": null,
    "pais": null,
    "data_inicio_atividade": "1967-06-30",
    "cnae_fiscal": 6204000,
    "cnae_fiscal_descricao": "Consultoria em tecnologia da informação",
    "descricao_tipo_de_logradouro": "AVENIDA",
    "logradouro": "L2 SGAN",
    "numero": "601",
    "complemento": "MODULO G",
    "bairro": "ASA NORTE",
    "cep": "70836900",
    "uf": "DF",
    "codigo_municipio": 9701,
    "codigo_municipio_ibge": 5300108,
    "municipio": "BRASILIA",
    "ddd_telefone_1": "",
    "ddd_telefone_2": "",
    "ddd_fax": "",
    "situacao_especial": "",
    "data_situacao_especial": null,
    "opcao_pelo_simples": null,
    "data_opcao_pelo_simples": null,
    "data_exclusao_do_simples": null,
    "opcao_pelo_mei": null,
    "data_opcao_pelo_mei": null,
    "data_exclusao_do_mei": null,
    "razao_social": "SERVICO FEDERAL DE PROCESSAMENTO DE DADOS (SERPRO)",
    "codigo_natureza_juridica": 2011,
    "natureza_juridica": "Empresa Pública",
    "qualificacao_do_responsavel": 16,
    "capital_social": 1061004800,
    "codigo_porte": 5,
    "porte": "DEMAIS",
    "ente_federativo_responsavel": null,
    "regime_tributario": null,
    "qsa": [
        {
            "identificador_de_socio": 2,
            "nome_socio": "ANDRE DE CESERO",
            "cnpj_cpf_do_socio": "***220050**",
            "codigo_qualificacao_socio": 10,
            "qualificacao_socio": "Diretor",
            "data_entrada_sociedade": "2016-06-16",
            "codigo_pais": null,
            "pais": null,
            "cpf_representante_legal": "***000000**",
            "nome_representante_legal": "",
            "codigo_qualificacao_representante_legal": 0,
            "qualificacao_representante_legal": null,
            "codigo_faixa_etaria": 6,
            "faixa_etaria": "Entre 51 a 60 anos"
        },
        {
            "identificador_de_socio": 2,
            "nome_socio": "ANTONIO DE PADUA FERREIRA PASSOS",
            "cnpj_cpf_do_socio": "***595901**",
            "codigo_qualificacao_socio": 10,
            "qualificacao_socio": "Diretor",
            "data_entrada_sociedade": "2016-12-08",
            "codigo_pais": null,
            "pais": null,
            "cpf_representante_legal": "***000000**",
            "nome_representante_legal": "",
            "codigo_qualificacao_representante_legal": 0,
            "qualificacao_representante_legal": null,
            "codigo_faixa_etaria": 7,
            "faixa_etaria": "Entre 61 a 70 anos"
        },
        {
            "identificador_de_socio": 2,
            "nome_socio": "WILSON BIANCARDI COURY",
            "cnpj_cpf_do_socio": "***414127**",
            "codigo_qualificacao_socio": 10,
            "qualificacao_socio": "Diretor",
            "data_entrada_sociedade": "2019-06-18",
            "codigo_pais": null,
            "pais": null,
            "cpf_representante_legal": "***000000**",
            "nome_representante_legal": "",
            "codigo_qualificacao_representante_legal": 0,
            "qualificacao_representante_legal": null,
            "codigo_faixa_etaria": 8,
            "faixa_etaria": "Entre 71 a 80 anos"
        },
        {
            "identificador_de_socio": 2,
            "nome_socio": "GILENO GURJAO BARRETO",
            "cnpj_cpf_do_socio": "***099595**",
            "codigo_qualificacao_socio": 16,
            "qualificacao_socio": "Presidente",
            "data_entrada_sociedade": "2020-02-03",
            "codigo_pais": null,
            "pais": null,
            "cpf_representante_legal": "***000000**",
            "nome_representante_legal": "",
            "codigo_qualificacao_representante_legal": 0,
            "qualificacao_representante_legal": null,
            "codigo_faixa_etaria": 5,
            "faixa_etaria": "Entre 41 a 50 anos"
        },
        {
            "identificador_de_socio": 2,
            "nome_socio": "RICARDO CEZAR DE MOURA JUCA",
            "cnpj_cpf_do_socio": "***989951**",
            "codigo_qualificacao_socio": 10,
            "qualificacao_socio": "Diretor",
            "data_entrada_sociedade": "2020-05-12",
            "codigo_pais": null,
            "pais": null,
            "cpf_representante_legal": "***000000**",
            "nome_representante_legal": "",
            "codigo_qualificacao_representante_legal": 0,
            "qualificacao_representante_legal": null,
            "codigo_faixa_etaria": 5,
            "faixa_etaria": "Entre 41 a 50 anos"
        },
        {
            "identificador_de_socio": 2,
            "nome_socio": "ANTONINO DOS SANTOS GUERRA NETO",
            "cnpj_cpf_do_socio": "***073447**",
            "codigo_qualificacao_socio": 5,
            "qualificacao_socio": "Administrador",
            "data_entrada_sociedade": "2019-02-11",
            "codigo_pais": null,
            "pais": null,
            "cpf_representante_legal": "***000000**",
            "nome_representante_legal": "",
            "codigo_qualificacao_representante_legal": 0,
            "qualificacao_representante_legal": null,
            "codigo_faixa_etaria": 7,
            "faixa_etaria": "Entre 61 a 70 anos"
        }
    ],
    "cnaes_secundarios": [
        {
            "codigo": 6201501,
            "descricao": "Desenvolvimento de programas de computador sob encomenda"
        },
        {
            "codigo": 6202300,
            "descricao": "Desenvolvimento e licenciamento de programas de computador customizáveis"
        },
        {
            "codigo": 6203100,
            "descricao": "Desenvolvimento e licenciamento de programas de computador não-customizáveis"
        },
        {
            "codigo": 6209100,
            "descricao": "Suporte técnico, manutenção e outros serviços em tecnologia da informação"
        },
        {
            "codigo": 6311900,
            "descricao": "Tratamento de dados, provedores de serviços de aplicação e serviços de hospedagem na internet"
        }
    ]
}
```

## Busca paginada

A busca paginada aceita um ou mais desses parâmetros na URL:

| Parâmetro | Descrição |
|---|---|
| `uf` | Sigla da UF com duas letras |
| `cnae_fiscal` | Código do CNAE fiscal |
| `cnae` | Busca o código tanto no CNAE fiscal como nos CNAES secundários |
| `cnpf` | Busca por CPF ou CNPJ da pessoa no quadro societário |
| `limit` | Número máximo de CNPJ por página (o máximo é 1.000) |
| `cursor` | Valor a ser passado para requisitar a próxima página da busca |

Por exemplo, a empresa do JSON anterior pode ser encontrada (bem como outras semelhantes) com: `GET /?uf=DF&cnae=6209100`.

Mais de um valor pode ser passado, seja repetindo o parâmetro, seja separando os valores por vírgulas. Por exemplo, para buscas no Rio Grande do Norte, Paraíba e Pernambuco, todas essas são opções válidas:

* `GET /?uf=rn&uf=pb&uf=pe`
* `GET /?uf=rn,pb,pe`
* `GET /?uf=rn,pb&uf=pe`

O mesmo vale para `cnae` e `cnae_fiscal`.

### Busca por CPF ou CNPJ da pessoa no quadro societário

Não utilizar pontuação ou barras nesses valores.

Para buscar por CPF, utilizar `*` como os três primeiros caracteres e como os dois últimos. Por exemplo, para buscar pelo CPF 123.456.789-01, utilizar `***456789**` — é assim que o CPF dos sócios aparece no banco de dados original.

### Exemplo de JSON de resposta:

```json
{
    "data": [],
    "cursor": "33683111000280"
}
```

`data` contém uma sequência de JSON como o do exemplo para uma única empresa.

#### Cursor

Com uma resposta dessas do exemplo, para requisitar a próxima página, basta adicionar `&cursor=33683111000280` ao final da URL.

Quando a resposta estievr sem `cursor`, isso significa que é a última página da busca.

## Dicionário de dados

A API web tenta manter o máximo de similaridade com o nome dos campos, significado e tipos de dados conforme os _layout_ divulgado pela Receita Federal, salvo em alguns casos:

* Campos com código numérico: adicionamos o “significado” do código numérico (por exemplo, `codigo_pais` é um número e `pais` é adicionado com o nome do país como texto)
* Campos _booleanos_: convertemos textos como `"S"` e `"N"` para valores `true`, `false` ou `null` (em branco)
* Código do município do IBGE é adicionado em `codigo_municipio_ibge`
* Dados em CSV relacionados são adicionados como _arrays_ (quadro societário, CNAEs secundários e regime tributário)

 Sobre o tipo dos dados:

* Todos os campos, salvo `cnpj` são opcionais e podem estar em branco
* Datas seguem o padrão `YYYY-MM-DD`

### Estrutura dos dados das empresas

#### Dados replicados dos originais

| Nome | Tipo | Origem | Nome no _layout_ da Receita Federal |
|---|---|---|---|
| `bairro` | `string` | `Estabelecimentos*.zip` | Bairro |
| `capital_social` | `number` | `Empresas*.zip` |  Capital social da empresa |
| `cep` | `string` | `Estabelecimentos*.zip` | CEP |
| `cnae_fiscal` | `string` | `Estabelecimentos*.zip` | CNAE fiscal principal |
| `codigo_municipio` | `string` | `Estabelecimentos*.zip` | Município |
| `codigo_natureza_juridica` | `number` | `Empresas*.zip` | Natureza jurídica |
| `codigo_pais` | `number` | `Estabelecimentos*.zip` | País |
| `codigo_porte` | `number` | `Empresas*.zip` |  Porte da empresa |
| `complemento` | `string` | `Estabelecimentos*.zip` | Complemeto |
| `data_exclusao_do_mei` | `string` | `Simples.zip` | Data de exclusão do MEI |
| `data_exclusao_do_simples` | `string` | `Simples.zip` | Data de exclusão do Simples |
| `data_inicio_atividade` | `string` | `Estabelecimentos*.zip` | Data de início atividade |
| `data_opcao_pelo_mei` | `string` | `Simples.zip` | Data de opção pelo MEI |
| `data_opcao_pelo_simples` | `string` | `Simples.zip` | Data de opção pelo Simples |
| `data_situacao_cadastral` | `string` | `Estabelecimentos*.zip` | Data situação cadastral |
| `data_situacao_especial` | `string` | `Estabelecimentos*.zip` | Data da situação especial |
| `descricao_tipo_de_logradouro` | `string` | `Estabelecimentos*.zip` | Tipo de logradouro |
| `ente_federativo_responsavel` | `string` | `Empresas*.zip` | Ente federativo responsável |
| `identificador_matriz_filial` | `number`  | `Estabelecimentos*.zip` | Identificador matriz/filial |
| `logradouro` | `string` | `Estabelecimentos*.zip` | Logradouro |
| `motivo_situacao_cadastral` | `numeric` | `Estabelicimentos*.zip` | Motivo da situação cadastral |
| `nome_cidade_no_exterior` | `string` | `Estabelicimentos*.zip` | Nome da cidade no exterior |
| `nome_fantasia` | `string` | `Estabelecimentos*.zip` | Nome fantasia |
| `numero` | `string` | `Estabelecimentos*.zip` | Número |
| `qualificacao_do_responsavel` | `number` | `Empresas*.zip` | Qualificação do responsável |
| `razao_social` | `string` | `Empresas*.zip` | Razão social / Nome empresarial |
| `situacao_cadastral` | `number` | `Estabelecimentos*.zip` | Situação cadastral |
| `situacao_especial` | `string` | `Estabelecimentos*.zip` | Situação especial |
| `uf` | `string` | `Estabelecimentos*.zip` | UF |

#### Dados enriquecidos

| Nome | Tipo | Origem | Descrição |
|---|---|---|---|
| `cnae_fiscal_descricao` | `string` | `Estabelecimentos*.zip` e `Cnaes.zip` | Conversão de acordo com arquivo `Cnaes.zip` |
| `cnpj` | `string` |  `Empresas*.zip` e `Estabelecimentos*.zip` | Concatenação de CNPJ Básico, CNPJ ordem e CNPJ DV |
| `codigo_municipio_ibge` | `number` | `Estabelecimentos*.zip` e `TABMUN.CSV` [do Tesouro Nacional](https://www.tesourotransparente.gov.br/ckan/dataset/abb968cb-3710-4f85-89cf-875c91b9c7f6/resource/eebb3bc6-9eea-4496-8bcf-304f33155282/) | Conversão de acordo com ambos os arquivos |
| `ddd_fax` | `string` | `Estabelecimentos*.zip` | Concatenação de DDD do fax e Fax |
| `ddd_telefone_1` | `string` | `Estabelecimentos*.zip` | Concatenação de DDD 1 e Telefone 1 |
| `ddd_telefone_2` | `string` | `Estabelecimentos*.zip` | Concatenação de DDD 2 e Telefone 2 |
| `descricao_identificador_matriz_filial` | `string` | `Estabelecimentos*.zip` | Conversão do `identificador_matriz_filial` de acordo com o _layout_ |
| `descricao_motivo_situacao_cadastral` | `string` | `Estabelecimentos*.zip` e `Motivos.zip` | Conversão de acordo com arquivo `Motivos.zip`  |
| `descricao_situacao_cadastral` | `string` | `Estabelecimentos*.zip` | Conversão da `situacao_cadastral` de acordo com o _layout_ |
| `municipio` | `string` | `Estabelecimentos*.zip` e `Municipios.zip` | Conversão de acordo com arquivo `Municipios.zip` |
| `natureza_juridica` | `string` | `Empresas*.zip` e `Naturezas.zip` | Conversão de acordo com arquivo `Naturezas.zip` |
| `opcao_pelo_mei` | `boolean` | `Simples.zip` | Conversão de `"S"`/`"N"` para `boolean` |
| `opcao_pelo_simples` | `boolean` | `Simples.zip` | Conversão de `"S"`/`"N"` para `boolean` |
| `pais` | `string` | `Estabelecimentos*.zip` e `Paises.zip` | Conversão de acordo com arquivo `Paises.zip` |
| `porte` | `string` | `Empresas*.zip` | Conversão de acordo com o _layout_ |

### Estrutura dos dados do quadro societário

#### Dados replicados dos originais

| Nome | Tipo | Origem | Nome no _layout_ da Receita Federal |
|---|---|---|---|
| `cnpj_cpf_do_socio` | `string` | `Socios*.zip` | CNPJ/CPF do sócio |
| `codigo_faixa_etaria` | `number` | `Socios*.zip` | Faixa etária |
| `codigo_pais` | `number` | `Socios*.zip` | País |
| `codigo_qualificacao_representante_legal` | `number` | `Socios*.zip` | Qualificação do representante legal |
| `codigo_qualificacao_socio` | `number` | `Socios*.zip` | Qualificação do sócio |
| `cpf_representante_legal` | `string` | `Socios*.zip` | Representante legal |
| `data_entrada_sociedade` | `string` | `Socios*.zip` | Data de entrada sociedade |
| `identificador_de_socio` | `number` | `Socios*.zip` | Identificador de sócio |
| `nome_representante_legal` | `string` | `Socios*.zip` | Nome do representante |
| `nome_socio` | `string` | `Socios*.zip` | Nome do sócio (no caso PF) ou razão social (no caso PJ) |

#### Dados enriquecidos

| Nome | Tipo | Origem | Descrição |
|---|---|---|---|
| `faixa_etaria` | `string` | `Socios*.zip` | Conversão de acordo com o _layout_ |
| `pais` | `string` | `Socios*.zip` e `Paises.zip` | Conversão de acordo com arquivo `Paises.zip` |
| `qualificacao_representante_legal` | `string` | `Socios*.zip` e `Qualificacoes.zip` | Conversão de acordo com arquivo `Qualificacoes.zip` |
| `qualificacao_socio` | `string` | `Socios*.zip` e `Qualificacoes.zip` | Conversão de acordo com arquivo `Qualificacoes.zip` |

### Estrutura dos dados dos CNAEs secundários

O campo `cnaes_secundarios` é um _array_ composto de:

#### Dados replicados dos originais

| Nome | Tipo | Origem | Nome no _layout_ da Receita Federal |
|---|---|---|---|
| `codigo` | `number` | `Estabelecimentos*.zip` | CNAE fiscal secundária |

#### Dados enriquecidos

| Nome | Tipo | Origem | Descrição |
|---|---|---|---|
| `descricao` | `string` | `Estabelecimentos*.zip` e `Cnaes.zip` | Conversao de acordo com arquivo `Cnaes.zip` |

### Estrutura dos dados do regime tributário

O campo `regime_tributário` é um _array_ com dados oriundos de `Imunes e Isentas.zip`, `Lucro Arbitrado.zip`, `Lucro Presumido.zip` ou `Lucro Real.zip` , composto de:

| Nome | Tipo |
|---|---|
| `ano` | `number` |
| `cnpj_da_scp` | `string` |
| `forma_de_tributação` | `string` |
| `quantidade_de_escrituracoes` | `number` |

## _Endpoints_ auxiliares

Para todos esses _endpoints_ é esperada resposta com status `200`:

| Caminho da URL | Tipo de requisição | Conteúdo esperado na resposta |
|---|---|---|
| `/updated` | `GET` | JSON contendo a data de extração dos dados pela Receita Federal. |
| `/healthz` | `GET` ou `HEAD` | Resposta sem conteúdo |
