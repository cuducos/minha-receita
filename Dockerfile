FROM golang:1.18-bullseye as build
WORKDIR /minha-receita
ADD go.* ./
ADD main.go .
ADD api/ ./api/
ADD cmd/ ./cmd/
ADD db/ ./db/
ADD download/ ./download/
ADD testdata/ ./testdata/
ADD transform/ ./transform/
ADD sample/ ./sample/
ADD check/ ./check/
RUN go get && go build -o /usr/bin/minha-receita

FROM debian:bullseye-slim
RUN apt-get update && \
    apt-get install -y --no-install-recommends postgresql-client ca-certificates && \
    update-ca-certificates && \
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /usr/bin/minha-receita /usr/bin/minha-receita
ENTRYPOINT ["/usr/bin/minha-receita"]
CMD ["api"]
