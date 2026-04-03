package cmd

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var simulatorURL string

var rootCmd = &cobra.Command{
	Use:   "sim-cli",
	Short: "NoTIP Simulator CLI",
	Long:  "A CLI tool for managing NoTIP Simulator gateways, sensors, and anomalies.",
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Disable PTerm styling when stdout is not a TTY (e.g., CI pipelines,
	// docker compose run without -it) to avoid raw ANSI escape sequences.
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		pterm.DisableStyling()
		pterm.DisableColor()
	}

	simulatorURL = os.Getenv("SIMULATOR_URL")
	if simulatorURL == "" {
		simulatorURL = "http://simulator:8090"
	}
}
