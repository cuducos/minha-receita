FROM python:3.7.5-slim

ENV PYTHONPATH=/api
WORKDIR /api
ADD Pipfile Pipfile
ADD Pipfile.lock Pipfile.lock

RUN apt-get update && \
    apt-get install -y gcc python3-dev && \
    python -m pip install --upgrade pip && \
    python -m pip install --upgrade black pipenv && \
    pipenv install --deploy --dev && \
    apt-get remove -y gcc python3-dev && \
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/*

ADD api/ ./api/

CMD ["pipenv", "run", "python", "-m", "api"]
