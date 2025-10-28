# Documentação

Utilizamos o [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/):

```console
$ docker pull squidfunk/mkdocs-material
$ docker run --rm -v $(pwd):/docs squidfunk/mkdocs-material build
```

A documentação vai ser gerada em `site/index.html`. Para servir enquanto desenvolve:

```console
$ docker run -p 8000:8000 --rm -v $(pwd):/docs squidfunk/mkdocs-material serve --dev-addr 0.0.0.0:8000
```
