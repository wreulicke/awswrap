package main

import (
	"os"

	"github.com/wurelicke/awswrap/cmd"
)

func main() {
	cmd := cmd.NewRootCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
