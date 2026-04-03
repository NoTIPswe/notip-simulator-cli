package cmd

import (
	"testing"

	"github.com/pterm/pterm"
)

func TestStartSpinner_RawOutputReturnsNoop(t *testing.T) {
	prevRaw := pterm.RawOutput
	pterm.RawOutput = true
	t.Cleanup(func() {
		pterm.RawOutput = prevRaw
	})

	sp := startSpinner("test")
	if _, ok := sp.(noopSpinner); !ok {
		t.Fatalf("expected noopSpinner in raw output mode, got %T", sp)
	}

	// No-op methods should be safe to call.
	sp.Success("ok")
	sp.Fail("err")
	sp.Warning("warn")
}

func TestPtermSpinner_NilInnerIsSafe(t *testing.T) {
	sp := ptermSpinner{}

	sp.Success("ok")
	sp.Fail("err")
	sp.Warning("warn")
}
