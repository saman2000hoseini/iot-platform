package main

import (
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/cmd"
	"os"
)

const (
	exitFailure = 1
)

func main() {
	root := cmd.NewRootCommand()

	if root != nil {
		if err := root.Execute(); err != nil {
			os.Exit(exitFailure)
		}
	}
}
