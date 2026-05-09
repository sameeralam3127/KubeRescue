FROM python:3.13-slim AS test

WORKDIR /app

COPY pyproject.toml README.md ./
COPY kuberescue ./kuberescue
COPY tests ./tests

RUN pip install --no-cache-dir -e ".[dev]"
RUN pytest -v && touch /tmp/tests-passed

FROM python:3.13-slim

WORKDIR /app

COPY pyproject.toml README.md ./
COPY kuberescue ./kuberescue
COPY --from=test /tmp/tests-passed /tmp/tests-passed

RUN pip install --no-cache-dir -e .

ENTRYPOINT ["kuberescue"]
CMD ["--help"]
