package main

import (
	"os"
	"testing"
)

func TestMain_HelpPath(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{"sim-cli", "--help"}
	t.Cleanup(func() {
		os.Args = oldArgs
	})

	main()
}
