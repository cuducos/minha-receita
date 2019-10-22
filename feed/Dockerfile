FROM python:3.7.5-slim

WORKDIR /feed
ADD Pipfile Pipfile
ADD Pipfile.lock Pipfile.lock

RUN echo "\n==> Updating apt\n" && \
    apt-get update && \
    echo "\n==> Installing PostgreSQL client\n" && \
    apt-get install -y ca-certificates gnupg2 lsb-release wget && \
    wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - && \
    sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt/ `lsb_release -cs`-pgdg main" >> /etc/apt/sources.list.d/pgdg.list' && \
    apt-get update && \
    apt-get install -y postgresql-client-11 && \
    echo "\n==> Installing build dependencies\n" && \
    apt-get install -y gcc libpq-dev python3-dev && \
    echo "\n==> Installing Python packages\n" && \
    python -m pip install --upgrade pip && \
    python -m pip install --upgrade black pipenv && \
    pipenv install --deploy --dev && \
    pipenv run python -m pip install --upgrade https://github.com/turicas/rows/archive/develop.zip && \
    echo "\n==> Cleaning up\n" && \
    apt-get remove -y ca-certificates gcc gnupg2 libpq-dev lsb-release python3-dev wget && \
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/*

ADD import.py import.py
CMD ["pipenv", "run", "python", "import.py"]
