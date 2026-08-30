package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sameeralam3127/kuberescue/internal/detect"
	"github.com/sameeralam3127/kuberescue/internal/diagnose"
	"github.com/sameeralam3127/kuberescue/internal/kube"
	"github.com/sameeralam3127/kuberescue/internal/report"
)

func newDiagnoseCommand(root *rootOptions) *cobra.Command {
	var (
		namespace string
		selector  string
		output    string
	)

	cmd := &cobra.Command{
		Use:   "diagnose [pod]",
		Short: "Explain why workloads are unhealthy without changing anything",
		Long: `Scan a namespace for CrashLoopBackOff, OOMKilled, ImagePullBackOff,
unschedulable (Pending), and stuck-rollout Deployments, and explain the
likely cause of each with the evidence behind it.

diagnose never mutates the cluster; it only reads pods, deployments, and
events. Pass a pod name to narrow the scan to just that pod.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if output != "text" && output != "json" {
				return fmt.Errorf("--output must be text or json, got %q", output)
			}

			client, err := kube.NewClient(root.kubeconfig, root.kubeContext)
			if err != nil {
				return err
			}

			opts := diagnose.Options{Namespace: namespace, Selector: selector}
			if len(args) == 1 {
				opts.Pod = args[0]
			}

			detectors := []detect.Detector{detect.CrashLoop{}, detect.OOMKilled{}, detect.ImagePull{}, detect.Pending{}}
			rolloutDetectors := []detect.RolloutDetector{detect.StuckRollout{}}

			r, err := diagnose.Run(cmd.Context(), client, detectors, rolloutDetectors, opts)
			if err != nil {
				return err
			}

			if output == "json" {
				err = report.DiagnoseJSON(os.Stdout, r)
			} else {
				err = report.DiagnoseText(os.Stdout, r)
			}
			if err != nil {
				return err
			}

			if r.Detected > 0 {
				return exitError{code: ExitFindings}
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "namespace to diagnose")
	cmd.Flags().StringVarP(&selector, "selector", "l", "", "label selector, for example app=api")
	cmd.Flags().StringVarP(&output, "output", "o", "text", "output format: text or json")

	return cmd
}
