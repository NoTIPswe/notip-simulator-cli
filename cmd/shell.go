package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Start an interactive shell session",
	Long: `Opens a persistent REPL where you can run multiple commands without
restarting the container. Type 'help' for available commands, 'exit' to quit.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		printWelcomeBanner()

		// Silence cobra's usage print on every error inside the REPL —
		// errors are shown by the individual commands themselves.
		rootCmd.SilenceUsage = true
		defer func() { rootCmd.SilenceUsage = false }()

		reader := bufio.NewReader(os.Stdin)
		for {
			printPrompt()

			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					fmt.Println()
					pterm.Info.Println("Goodbye!")
				}
				return nil
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if line == "exit" || line == "quit" {
				pterm.Info.Println("Goodbye!")
				return nil
			}

			// Prevent the user from nesting shells.
			parts := strings.Fields(line)
			if parts[0] == "shell" {
				pterm.Warning.Println("Already inside a shell session.")
				continue
			}

			rootCmd.SetArgs(parts)
			if execErr := rootCmd.Execute(); execErr != nil {
				pterm.Error.Println(execErr)
			}
		}
	},
}

func printPrompt() {
	if pterm.RawOutput {
		fmt.Print("sim-cli> ")
	} else {
		fmt.Print(pterm.Green("sim-cli") + pterm.Gray(" › ") + " ")
	}
}

func printWelcomeBanner() {
	if !pterm.RawOutput {
		if err := pterm.DefaultBigText.WithLetters(
			putils.LettersFromStringWithStyle("sim", pterm.NewStyle(pterm.FgGreen)),
			putils.LettersFromStringWithStyle("-cli", pterm.NewStyle(pterm.FgGray)),
		).Render(); err != nil {
			pterm.Warning.Printf("failed to render banner: %v\n", err)
		}
	}
	pterm.DefaultBox.
		WithTitle("Interactive Shell").
		WithTitleTopCenter().
		Println("Type a command to execute it.\n" +
			pterm.Gray("  gateways list\n") +
			pterm.Gray("  gateways get <uuid>\n") +
			pterm.Gray("  sensors list <id>\n") +
			pterm.Gray("  anomalies disconnect <uuid> --duration 5\n") +
			"\nType " + pterm.Bold.Sprint("exit") + " or press Ctrl+D to quit.")
}

func init() {
	rootCmd.AddCommand(shellCmd)
}
