# Contributing

Thanks for helping improve KubeRescue.

## Local Checks

Before opening a pull request, run:

```bash
black --check .
ruff check .
flake8 .
mypy kuberescue
bandit -r kuberescue
pytest -v
```

## Pull Request Guidelines

- Keep changes focused and easy to review.
- Add or update tests for behavior changes.
- Update the README or docs when commands, setup, or behavior changes.
- Avoid adding new dependencies unless they clearly reduce project risk or complexity.
