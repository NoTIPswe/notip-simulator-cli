package main

import (
	"os"

	"github.com/NoTIPswe/notip-simulator-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
