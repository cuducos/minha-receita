FROM golang:1.15.3-buster as build
WORKDIR /minha-receita-api
ADD . .
RUN go get && go build -o /usr/bin/minha-receita-api

FROM debian:buster-slim
COPY --from=build /usr/bin/minha-receita-api /usr/bin/minha-receita-api
CMD ["/usr/bin/minha-receita-api"]
