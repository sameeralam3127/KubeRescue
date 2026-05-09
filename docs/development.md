# Development

## Setup

```bash
python -m venv .venv
source .venv/bin/activate
pip install -e ".[dev]"
```

## Quality Checks

```bash
black --check .
ruff check .
flake8 .
mypy kuberescue
bandit -r kuberescue
pytest -v
```

## Docker Image

```bash
docker build -t kuberescue:local .
docker run --rm kuberescue:local --help
```
