package main

import (
	"os"

	"github.com/andreagrandi/logbasset/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
