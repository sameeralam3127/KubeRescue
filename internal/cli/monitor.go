package cli

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/sameeralam3127/kuberescue/internal/detect"
	"github.com/sameeralam3127/kuberescue/internal/engine"
	"github.com/sameeralam3127/kuberescue/internal/kube"
	"github.com/sameeralam3127/kuberescue/internal/report"
)

func newMonitorCommand(root *rootOptions) *cobra.Command {
	var (
		namespace   string
		selector    string
		interval    time.Duration
		once        bool
		dryRun      bool
		maxRestarts int
		output      string
	)

	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Watch a namespace for failing pods and remediate them",
		Long: `Watch a namespace for pods in CrashLoopBackOff and restart them by
deleting the pod so its controller recreates it.

Pods without a controller are never deleted. Use --dry-run to preview
actions without changing the cluster.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if output != "text" && output != "json" {
				return fmt.Errorf("--output must be text or json, got %q", output)
			}

			logger := slog.New(logHandler(root.logLevel))
			client, err := kube.NewClient(root.kubeconfig, root.kubeContext)
			if err != nil {
				return err
			}

			eng := &engine.Engine{
				Client:    client,
				Detectors: []detect.Detector{detect.CrashLoop{}},
				Logger:    logger,
			}
			opts := engine.Options{
				Namespace:   namespace,
				Selector:    selector,
				DryRun:      dryRun,
				MaxRestarts: maxRestarts,
			}

			publish := func(r *engine.Report) {
				var err error
				if output == "json" {
					err = report.JSON(os.Stdout, r)
				} else {
					err = report.Text(os.Stdout, r)
				}
				if err != nil {
					logger.Error("writing report", "error", err)
				}
			}

			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			if once {
				r, err := eng.Scan(ctx, opts)
				if err != nil {
					return err
				}
				publish(r)
				if r.Detected > 0 {
					return exitError{code: ExitFindings}
				}
				return nil
			}

			logger.Info("monitoring namespace",
				"namespace", namespace,
				"interval", interval,
				"dryRun", dryRun,
			)
			if err := eng.Monitor(ctx, opts, interval, publish); err != nil && ctx.Err() == nil {
				return err
			}
			logger.Info("shutting down")
			return nil
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "namespace to monitor")
	cmd.Flags().StringVarP(&selector, "selector", "l", "", "label selector, for example app=api")
	cmd.Flags().DurationVarP(&interval, "interval", "i", 30*time.Second, "time between scans, for example 30s or 2m")
	cmd.Flags().BoolVar(&once, "once", false, "scan once and exit (exit code 2 when findings exist)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "report what would be done without changing the cluster")
	cmd.Flags().IntVar(&maxRestarts, "max-restarts", 0, "maximum pods to restart per scan (0 = unlimited)")
	cmd.Flags().StringVarP(&output, "output", "o", "text", "output format: text or json")

	return cmd
}
