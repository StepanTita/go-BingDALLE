package main

import (
	"os"

	"github.com/StepanTita/go-BingDALLE/cli"
)

func main() {
	if !cli.Run(os.Args) {
		os.Exit(2)
	}
	os.Exit(0)
}
