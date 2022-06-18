# build minha-receita bin
FROM ghcr.io/cuducos/minha-receita:main
WORKDIR /minha-receita
RUN apt-get update && \
    apt-get install -y --no-install-recommends screen && \
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/*
CMD ["./update.sh"]
