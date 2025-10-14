# Dicionário de dados

A API web tenta manter o máximo de similaridade com o nome dos campos, significado e tipos de dados conforme os _layout_ divulgado pela Receita Federal, salvo em alguns casos:

* Campos com código numérico: adicionamos o “significado” do código numérico (por exemplo, `codigo_pais` é um número e `pais` é adicionado com o nome do país como texto)
* Campos _booleanos_: convertemos textos como `"S"` e `"N"` para valores `true`, `false` ou `null` (em branco)
* Código do município do IBGE é adicionado em `codigo_municipio_ibge`
* Dados em CSV relacionados são adicionados como _arrays_ (quadro societário, CNAEs secundários e regime tributário)

 Sobre o tipo dos dados:

* Todos os campos, salvo `cnpj` são opcionais e podem estar em branco
* Datas seguem o padrão `YYYY-MM-DD`

## Estrutura dos dados das empresas

### Dados replicados dos originais


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


### Dados enriquecidos

!!! warning "Aviso"
    A Receita Federal, em outubro de 2025, passou a divulgar diversos códigos de países que não constam no `Paises.zip`. Seguindo outros documentos da própria Receita Federal ([exemplo](https://balanca.economia.gov.br/balanca/bd/tabelas/PAIS.csv)) o Minha Receita vai adicionar esses códigos de países nos dados: 15 (Aland, Ilhas), 150 (Ilhas do Guernsey), 151 (Canárias, Ilhas), 200 (Curaçao), 321 (Guernsey), 359 (Ilha de Man), 367 (Inglaterra), 393 (Jersey), 449 (Macedônia), 452 (Madeira, Ilha da), 498 (Montenegro), 578 (Palestina), 678 (Saint Kitts e Nevis), 699 (Sint Maarten), 737 (Sérvia) e 994 (A Designar).


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


## Estrutura dos dados do quadro societário

### Dados replicados dos originais


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


### Dados enriquecidos


| Nome | Tipo | Origem | Descrição |
|---|---|---|---|
| `faixa_etaria` | `string` | `Socios*.zip` | Conversão de acordo com o _layout_ |
| `pais` | `string` | `Socios*.zip` e `Paises.zip` | Conversão de acordo com arquivo `Paises.zip` |
| `qualificacao_representante_legal` | `string` | `Socios*.zip` e `Qualificacoes.zip` | Conversão de acordo com arquivo `Qualificacoes.zip` |
| `qualificacao_socio` | `string` | `Socios*.zip` e `Qualificacoes.zip` | Conversão de acordo com arquivo `Qualificacoes.zip` |


## Estrutura dos dados dos CNAEs secundários

O campo `cnaes_secundarios` é um _array_ composto de:

### Dados replicados dos originais


| Nome | Tipo | Origem | Nome no _layout_ da Receita Federal |
|---|---|---|---|
| `codigo` | `number` | `Estabelecimentos*.zip` | CNAE fiscal secundária |


### Dados enriquecidos


| Nome | Tipo | Origem | Descrição |
|---|---|---|---|
| `descricao` | `string` | `Estabelecimentos*.zip` e `Cnaes.zip` | Conversao de acordo com arquivo `Cnaes.zip` |


## Estrutura dos dados do regime tributário

O campo `regime_tributário` é um _array_ com dados oriundos de `Imunes e Isentas.zip`, `Lucro Arbitrado.zip`, `Lucro Presumido.zip` ou `Lucro Real.zip` , composto de:

| Nome | Tipo |
|---|---|
| `ano` | `number` |
| `cnpj_da_scp` | `string` |
| `forma_de_tributação` | `string` |
| `quantidade_de_escrituracoes` | `number` |
