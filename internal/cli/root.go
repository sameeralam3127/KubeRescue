// Package cli implements the kuberescue command line interface.
//
// Exit codes: 0 means no unhealthy pods were found, 1 means a runtime or
// usage error occurred, 2 means unhealthy pods were found (single-scan
// mode only), so CI pipelines can gate on findings.
package cli

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

// Exit codes returned by Execute.
const (
	ExitOK       = 0
	ExitError    = 1
	ExitFindings = 2
)

// version is injected at build time via -ldflags.
var version = "dev"

type rootOptions struct {
	kubeconfig  string
	kubeContext string
	logLevel    string
}

// Execute runs the CLI and returns the process exit code.
func Execute() int {
	opts := &rootOptions{}

	root := &cobra.Command{
		Use:           "kuberescue",
		Short:         "Kubernetes remediation engine: detect, diagnose, and safely fix failing workloads",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringVar(&opts.kubeconfig, "kubeconfig", "", "path to kubeconfig (defaults to standard loading rules)")
	root.PersistentFlags().StringVar(&opts.kubeContext, "context", "", "kubeconfig context to use")
	root.PersistentFlags().StringVar(&opts.logLevel, "log-level", "info", "log level: debug, info, warn, error")

	root.AddCommand(newMonitorCommand(opts))
	root.AddCommand(newDiagnoseCommand(opts))

	code := ExitOK
	root.SetArgs(os.Args[1:])
	if err := root.Execute(); err != nil {
		if fe, ok := err.(exitError); ok {
			return fe.code
		}
		slog.New(logHandler(opts.logLevel)).Error("command failed", "error", err)
		return ExitError
	}
	return code
}

// exitError carries a specific exit code out of a command without printing
// anything further.
type exitError struct{ code int }

func (e exitError) Error() string { return "" }

func logHandler(level string) slog.Handler {
	var l slog.Level
	switch level {
	case "debug":
		l = slog.LevelDebug
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}
	return slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: l})
}
