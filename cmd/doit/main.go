package main

import (
	"os"

	"github.com/Actual-Outcomes/doit/internal/cli"
)

func main() {
	if err := cli.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
