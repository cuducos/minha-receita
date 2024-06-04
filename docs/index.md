# Sobre

![Minha Receita](minha-receita.svg)

## Histórico

Pela [Lei de Acesso à Informação](http://www.acessoainformacao.gov.br/assuntos/conheca-seu-direito/a-lei-de-acesso-a-informacao), os dados de CNPJ devem ser públicos e acessíveis por máquina. A Receita Federal oferece esses dados escondidos atrás de um CAPTCHA ou espalhados em diversos arquivos CSV separados, dificultando a consulta de todas as informações sobre um CNPJ qualquer.

## Propósito

O código desse repositório faz esses dados ainda mais acessíveis:

1. Consolidar todas as informações em um banco de dados único
1. Fornecendo uma API web para a consulta de dados de um CNPJ

## Como usar

Disponibilizo essa aplicação para que cada um rode na sua própria infraestrutura, mas existe um protótipo no ar em [`https://minhareceita.org/`](https://minhareceita.org) (confira o [monitor de status](https://stats.uptimerobot.com/tqpD6AQZqI)).

Para fazer uma consulta usando o protótipo, acrescente o CNPJ a ser consultado ao final da URL. Por exemplo: `https://minhareceita.org/33.683.111/0002-80`. Para mais detalhes sobre como utilizar a API, confira a seção [Como usar](como-usar.md).

### Limites do protótipo

O protótipo não tem nenhuma [garantia de nível de serviço](https://pt.wikipedia.org/wiki/Acordo_de_n%C3%ADvel_de_servi%C3%A7o) e a única forma de aumentar sua disponibilidade é com [contribuições mensais](https://github.com/sponsors/cuducos) ou contribuições pontuais com Bitcoin (`13WCAR21g1LGqzzn6WTNV5g7QdN1J35BDk`) e Pix (`d6ede813-6621-4df4-9a93-8d0108fd9b6a`).

Mais sobre o protótipo nesse [fio no Twitter](https://twitter.com/cuducos/status/1339980776985808901).

Não tenho interesse em desenvolver um sistema para cobrar por esse serviço.

## Contato

Para conversar sobre o projeto, prefira [abrir uma _issue_ no GitHub](https://github.com/cuducos/minha-receita/issues/new) ou iniciar uma conversa pública [Mastodon](https://mastodon.social/@cuducos) ou [Bluesky](https://bsky.app/profile/cuducos.me). **Não responderei** mensagens deixadas como _DM_ ou emails em minhas contas pessoais:

* Esse é um projeto de dados abertos e código aberto, não existe motivo para conversas privadas relacionadas ao projeto
* Sua dúvida pode ser a de outra pessoa, e ter a nossa conversa em umas dessas três plataformas faz com que outras pessoas (que talvez tenham a mesma dúvida que você, ou dúvidas semelhantes) possam encontrar a nossa conversa
* Minha resposta pode ser incompleta ou mesmo errada, e ter essas conversas em ambiente aberto possibilitam que outras pessoas te ajudem, me corrijam e complementem o conteúdo
* Por fim, pode ser que tua dúvida já tenha sido respondida e, caso você não tenha encontrado, eu posso te enviar unm _link_ e você se junta a uma conversa que já está em andamento sobre o mesmo tema

## Muito obrigado

Ao [Turicas](https://twitter.com/turicas) por todo ativismo, mais o trabalho de coleta quando o formato dos arquivos ainda não era em CSV. Muito desse projeto se deve a ele. Ao [Bruno](https://twitter.com/555112299jedi), sem o qual [nunca teríamos acesso a esses dados por menos de R$ 500 mil](https://medium.com/serenata/o-dia-que-a-receita-nos-mandou-pagar-r-500-mil-para-ter-dados-p%C3%BAblicos-8e18438f3076). E ao [Fireman](https://twitter.com/daniellfireman), pela mentoria em Go!
