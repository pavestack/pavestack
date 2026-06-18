package main

import (
	"os"

	"github.com/pavestack/pave/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
