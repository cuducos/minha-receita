FROM golang:1.15.3-buster as BUILD
WORKDIR /minha-receita
ADD go.* ./
ADD main.go .
ADD api/ ./api/
ADD cmd/ ./cmd/
ADD db/ ./db/
ADD download/ ./download/
ADD testdata/ ./testdata/
ADD transform/ ./transform/
RUN go get && go test ./... && go build -o /usr/bin/minha-receita

FROM debian:buster-slim as FINAL
RUN apt-get update && \
    apt-get install -y postgresql-client && \
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/*

COPY --from=BUILD /usr/bin/minha-receita /usr/bin/minha-receita
COPY --from=BUILD /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/usr/bin/minha-receita"]
CMD ["api"]
