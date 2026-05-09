# Project Structure

```text
.
├── deploy/kubernetes/     Kubernetes manifests for local or staging installs
├── docs/                  Maintainer and contributor documentation
├── examples/              Demo workloads for manual testing
├── kuberescue/             Python package source
├── tests/                 Unit tests
├── Dockerfile             Container build with test stage
├── pyproject.toml         Python package and tool configuration
└── README.md              User-facing quick start
```

The project is intentionally small. New directories should be added only when
they support a real workflow such as packaging, deployment, examples, or docs.
