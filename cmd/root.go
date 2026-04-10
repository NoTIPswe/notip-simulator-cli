package cmd

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

// resetAllCommandFlags clears flag values/changed state across the command tree.
// Cobra flag state is sticky within the same process, which affects shell mode.
func resetAllCommandFlags(c *cobra.Command) {
	c.Flags().VisitAll(func(f *pflag.Flag) {
		if slice, ok := f.Value.(interface{ Replace([]string) error }); ok {
			_ = slice.Replace(nil)
		} else {
			_ = f.Value.Set(f.DefValue)
		}
		f.Changed = false
	})
	for _, child := range c.Commands() {
		resetAllCommandFlags(child)
	}
}

func init() {
	// Disable PTerm styling when stdout is not a TTY (e.g., CI pipelines, redirected output).
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		pterm.DisableStyling()
		pterm.DisableColor()
	}

	simulatorURL = os.Getenv("SIMULATOR_URL")
	if simulatorURL == "" {
		simulatorURL = "http://simulator:8090"
	}
}
