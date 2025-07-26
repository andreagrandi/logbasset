package main

import (
	"os"

	"github.com/andreagrandi/logbasset/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
