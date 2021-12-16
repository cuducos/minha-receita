# Sobre

![Minha Receita](minha-receita.svg)

## Histórico

Pela [Lei de Acesso à Informação](http://www.acessoainformacao.gov.br/assuntos/conheca-seu-direito/a-lei-de-acesso-a-informacao), os dados de CNPJ devem ser públicos e acessíveis por máquina. A Receita Federal oferece esses dados escondidos atrás de um CAPTCHA ou espalhados em diversos arquivos CSV separados, dificultando a consulta de todas as informações sobre um CNPJ qualquer.

## Propósito

O código desse repositório faz esses dados ainda mais acessíveis:

1. Consolidar todas as informações de um determinado CNPJ em um arquivo único (no caso, um JSON por CNPJ)
1. Importando automaticamente os dados para um banco de dados PostgreSQL
1. Fornecendo uma API web para a consulta de dados de um CNPJ

## Qual a URL para acesso?

Disponibilizo essa aplicação para que cada um rode na sua própria infraestrutura, mas existe um protótipo no ar em [`https://minhareceita.org/`](https://minhareceita.org).

O protótipo não tem nenhuma [garantia de nível de serviço](https://pt.wikipedia.org/wiki/Acordo_de_n%C3%ADvel_de_servi%C3%A7o) e a única forma de aumentar sua disponibilidade é contribuindo via [financiamento coletivo no GitHub](https://github.com/sponsors/cuducos) ou Bitcoin (`13WCAR21g1LGqzzn6WTNV5g7QdN1J35BDk`).

Mais sobre o protótipo nesse [fio no Twitter](https://twitter.com/cuducos/status/1339980776985808901).

Não tenho interesse em desenvolver um sistema para cobrar por esse serviço.

## Muito obrigado

Ao [Turicas](https://twitter.com/turicas) por todo ativismo, mais o trabalho de coleta quando o formato dos arquivos ainda não era em CSV. Muito desse projeto se deve a ele. Ao [Bruno](https://twitter.com/555112299jedi), sem o qual [nunca teríamos acesso a esses dados por menos de R$ 500 mil](https://medium.com/serenata/o-dia-que-a-receita-nos-mandou-pagar-r-500-mil-para-ter-dados-p%C3%BAblicos-8e18438f3076). E ao [Fireman](https://twitter.com/daniellfireman), pela mentoria em Go!
