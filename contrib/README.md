# Relatório sobre os dados abertos de CNPJ da Receita Federal

Esse material é escrito em Python pela facilidade de evoluir código e relatório em um local só com o Jupyter Notebook.

O propósito é gerar automaticamente um relatório com inconsistência nos dados abertos utilizados pelo projeto como [datas incorretas](https://twitter.com/cuducos/status/1479078346248097793) e [informações aparentemente faltantes ou inconsistentes](https://github.com/cuducos/minha-receita/issues/37#issuecomment-1006069395).

## Hipóteses

* Todas as datas devem ser válidas e no passado
* Todo estabelecimento (`ESTABELE`) deve ter uma correlação nos dados da base do CNPJ (`EMPRECSV`)
* Todo CNPJ base deve ter ao menos 1 estabelecimento
* Toda entrada do quadro societário (`SOCIOCSV`) deve ter uma correlação nos dados da base do CNPJ (`EMPRECSV`)
* Toda entrada de enquadramento no Simples ou MEI (`SIMPES`) deve ter uma correlação nos dados da base do CNPJ

## Instalação

* Requer Python 3 (testado manualmente com Python 3.9) e dependências para instalação do NumPy
* Fazer o download dos dados (por exemplo, com `go run main.go download`)
* Descompactar os arquivos com `python contrib/unzip.py data/` (considerando que `data/` é onde estão os arquivos)
* Em um ambiente virtual, instale os pacotes Python com `pip install -r contrib/requirements.txt`
* Inicie o Jupyter com `jupyter notebook`
* Na interface do navegador, acesse `contrib/report.ipynb`
