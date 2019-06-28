FROM python:3.7.3-slim

WORKDIR /feed
ADD Pipfile Pipfile
ADD Pipfile.lock Pipfile.lock

RUN apt-get update && \
    apt-get install -y gnupg2 ca-certificates lsb-release wget && \
    wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - && \
    sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt/ `lsb_release -cs`-pgdg main" >> /etc/apt/sources.list.d/pgdg.list' && \
    apt-get update && \
    apt-get install -y gcc libpq-dev python3-dev postgresql-client-11 && \
    python -m pip install --upgrade pip && \
    python -m pip install --upgrade black pipenv && \
    pipenv install --system --dev && \
    python -m pip install --upgrade https://github.com/turicas/rows/archive/develop.zip

ADD import.sh import.sh
ADD import_cnae.py import_cnae.py
