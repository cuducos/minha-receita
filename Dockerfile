FROM golang:1.21-bookworm AS build
WORKDIR /minha-receita
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o /usr/bin/minha-receita

FROM debian:bookworm-slim
LABEL org.opencontainers.image.description="Sua API web para consulta de informações do CNPJ da Receita Federal"
LABEL org.opencontainers.image.source="https://github.com/cuducos/minha-receita"
LABEL org.opencontainers.image.title="Minha Receita"

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    update-ca-certificates && \
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /usr/bin/minha-receita /usr/bin/minha-receita
ENTRYPOINT ["/usr/bin/minha-receita"]
CMD ["api"]
