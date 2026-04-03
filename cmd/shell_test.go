package cmd

import (
	"testing"

	"github.com/pterm/pterm"
)

func TestPrintPrompt_RawOutput(t *testing.T) {
	prevRaw := pterm.RawOutput
	pterm.RawOutput = true
	t.Cleanup(func() {
		pterm.RawOutput = prevRaw
	})

	out := captureStdout(t, printPrompt)
	if out != "sim-cli> " {
		t.Fatalf("unexpected prompt: %q", out)
	}
}

func TestPrintWelcomeBanner_RawOutput(t *testing.T) {
	prevRaw := pterm.RawOutput
	pterm.RawOutput = true
	t.Cleanup(func() {
		pterm.RawOutput = prevRaw
	})

	_ = captureStdout(t, printWelcomeBanner)
}
