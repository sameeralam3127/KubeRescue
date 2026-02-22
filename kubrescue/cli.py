import typer
from kubrescue.watcher.monitor import monitor_namespace

app = typer.Typer(help="KubeRescue - Kubernetes Auto Remediation Engine")

@app.command()
def monitor(namespace: str = "default"):
    """
    Monitor a namespace for failing pods and remediate.
    """
    monitor_namespace(namespace)


if __name__ == "__main__":
    app()