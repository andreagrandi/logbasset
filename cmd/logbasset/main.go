package main

import (
	"github.com/andreagrandi/logbasset/internal/cli"
	"github.com/andreagrandi/logbasset/internal/errors"
)

func main() {
	err := cli.Execute()
	errors.HandleErrorAndExit(err)
}
