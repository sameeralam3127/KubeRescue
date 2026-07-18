package main

import (
	"os"

	"github.com/sameeralam3127/kuberescue/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
