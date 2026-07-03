import typer
from typing import Annotated
from kuberescue.watcher.monitor import monitor_namespace

app = typer.Typer(help="KubeRescue - Kubernetes Auto Remediation Engine")


@app.command()
def monitor(
    namespace: Annotated[
        str,
        typer.Option(
            "--namespace",
            "-n",
            help="Kubernetes namespace to monitor.",
        ),
    ] = "default",
    interval: Annotated[
        int,
        typer.Option(
            "--interval",
            "-i",
            min=1,
            help="Seconds to wait between scans.",
        ),
    ] = 10,
    once: Annotated[
        bool,
        typer.Option(
            "--once",
            help="Run one scan and exit.",
        ),
    ] = False,
    dry_run: Annotated[
        bool,
        typer.Option(
            "--dry-run",
            help="Report matching pods without deleting them.",
        ),
    ] = False,
    max_restarts: Annotated[
        int | None,
        typer.Option(
            "--max-restarts",
            min=1,
            help="Maximum pods to restart per scan.",
        ),
    ] = None,
) -> None:
    """
    Monitor a namespace for failing pods and remediate.
    """
    monitor_namespace(
        namespace,
        interval_seconds=interval,
        once=once,
        dry_run=dry_run,
        max_restarts=max_restarts,
    )


if __name__ == "__main__":
    app()
