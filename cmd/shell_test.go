package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/pterm/pterm"
	"golang.org/x/term"
)

type shellReadEvent struct {
	line string
	err  error
}

type scriptedShellEditor struct {
	events []shellReadEvent
	index  int
}

func (s *scriptedShellEditor) ReadLine() (string, error) {
	if s.index >= len(s.events) {
		return "", io.EOF
	}
	event := s.events[s.index]
	s.index++
	return event.line, event.err
}

func setShellHooksForTest(t *testing.T) {
	t.Helper()

	prevStdin := shellStdin
	prevStdout := shellStdout
	prevIsTerminal := shellIsTerminal
	prevMakeRaw := shellMakeRaw
	prevRestore := shellRestore
	prevNewEditor := shellNewEditor
	prevRenderBanner := renderWelcomeBigText

	t.Cleanup(func() {
		shellStdin = prevStdin
		shellStdout = prevStdout
		shellIsTerminal = prevIsTerminal
		shellMakeRaw = prevMakeRaw
		shellRestore = prevRestore
		shellNewEditor = prevNewEditor
		renderWelcomeBigText = prevRenderBanner
	})
}

func makePipePair(t *testing.T) (*os.File, *os.File) {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	t.Cleanup(func() {
		_ = r.Close()
		_ = w.Close()
	})
	return r, w
}

func TestPrintPromptRawOutput(t *testing.T) {
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

func TestPrintWelcomeBannerRawOutput(t *testing.T) {
	prevRaw := pterm.RawOutput
	pterm.RawOutput = true
	t.Cleanup(func() {
		pterm.RawOutput = prevRaw
	})

	_ = captureStdout(t, printWelcomeBanner)
}

func TestPrintPromptNonRawOutput(t *testing.T) {
	prevRaw := pterm.RawOutput
	pterm.RawOutput = false
	t.Cleanup(func() {
		pterm.RawOutput = prevRaw
	})

	out := captureStdout(t, printPrompt)
	if !strings.Contains(out, "sim-cli") {
		t.Fatalf("unexpected prompt: %q", out)
	}
}

func TestPrintWelcomeBannerNonRawOutput(t *testing.T) {
	prevRaw := pterm.RawOutput
	pterm.RawOutput = false
	t.Cleanup(func() {
		pterm.RawOutput = prevRaw
	})

	_ = captureStdout(t, printWelcomeBanner)
}

func TestPrintWelcomeBannerRenderError(t *testing.T) {
	setShellHooksForTest(t)

	prevRaw := pterm.RawOutput
	pterm.RawOutput = false
	t.Cleanup(func() {
		pterm.RawOutput = prevRaw
	})

	renderWelcomeBigText = func() error {
		return errors.New("render failed")
	}

	_ = captureStdout(t, printWelcomeBanner)
}

func TestRunShellWithLineEditorEOF(t *testing.T) {
	setShellHooksForTest(t)

	inR, _ := makePipePair(t)
	_, outW := makePipePair(t)
	shellStdin = func() *os.File { return inR }
	shellStdout = func() *os.File { return outW }
	shellMakeRaw = func(int) (*term.State, error) { return &term.State{}, nil }
	shellRestore = func(int, *term.State) error { return nil }
	shellNewEditor = func(io.ReadWriter, string) shellLineEditor {
		return &scriptedShellEditor{events: []shellReadEvent{{err: io.EOF}}}
	}

	if err := runShellWithLineEditor(); err != nil {
		t.Fatalf("runShellWithLineEditor() error = %v, want nil", err)
	}
}

func TestRunShellWithLineEditorReadError(t *testing.T) {
	setShellHooksForTest(t)

	inR, _ := makePipePair(t)
	_, outW := makePipePair(t)
	shellStdin = func() *os.File { return inR }
	shellStdout = func() *os.File { return outW }
	shellMakeRaw = func(int) (*term.State, error) { return &term.State{}, nil }
	shellRestore = func(int, *term.State) error { return nil }
	shellNewEditor = func(io.ReadWriter, string) shellLineEditor {
		return &scriptedShellEditor{events: []shellReadEvent{{err: errors.New("read failed")}}}
	}

	if err := runShellWithLineEditor(); err == nil {
		t.Fatal("expected read error, got nil")
	}
}

func TestRunShellWithLineEditorRestoreError(t *testing.T) {
	setShellHooksForTest(t)

	inR, _ := makePipePair(t)
	_, outW := makePipePair(t)
	shellStdin = func() *os.File { return inR }
	shellStdout = func() *os.File { return outW }
	shellMakeRaw = func(int) (*term.State, error) { return &term.State{}, nil }
	restoreCalls := 0
	shellRestore = func(int, *term.State) error {
		restoreCalls++
		if restoreCalls == 1 {
			return errors.New("restore failed")
		}
		return nil
	}
	shellNewEditor = func(io.ReadWriter, string) shellLineEditor {
		return &scriptedShellEditor{events: []shellReadEvent{{line: ""}}}
	}

	if err := runShellWithLineEditor(); err == nil {
		t.Fatal("expected restore error, got nil")
	}
}

func TestRunShellWithLineEditorMakeRawAfterCommandError(t *testing.T) {
	setShellHooksForTest(t)

	inR, _ := makePipePair(t)
	_, outW := makePipePair(t)
	shellStdin = func() *os.File { return inR }
	shellStdout = func() *os.File { return outW }
	makeRawCalls := 0
	shellMakeRaw = func(int) (*term.State, error) {
		makeRawCalls++
		if makeRawCalls == 2 {
			return nil, errors.New("make raw failed")
		}
		return &term.State{}, nil
	}
	shellRestore = func(int, *term.State) error { return nil }
	shellNewEditor = func(io.ReadWriter, string) shellLineEditor {
		return &scriptedShellEditor{events: []shellReadEvent{{line: ""}}}
	}

	if err := runShellWithLineEditor(); err == nil {
		t.Fatal("expected make-raw error, got nil")
	}
}

func TestShellCommandUsesLineEditor(t *testing.T) {
	setShellHooksForTest(t)

	inR, _ := makePipePair(t)
	_, outW := makePipePair(t)
	shellStdin = func() *os.File { return inR }
	shellStdout = func() *os.File { return outW }
	shellIsTerminal = func(int) bool { return true }
	shellMakeRaw = func(int) (*term.State, error) { return &term.State{}, nil }
	shellRestore = func(int, *term.State) error { return nil }
	shellNewEditor = func(io.ReadWriter, string) shellLineEditor {
		return &scriptedShellEditor{events: []shellReadEvent{{line: "exit"}}}
	}

	if err := runCmd("shell"); err != nil {
		t.Fatalf("shell command failed with line editor: %v", err)
	}
}

func TestShellCommandFallsBackWhenLineEditorFails(t *testing.T) {
	setShellHooksForTest(t)

	inR, inW := makePipePair(t)
	_, outW := makePipePair(t)
	shellStdin = func() *os.File { return inR }
	shellStdout = func() *os.File { return outW }
	shellIsTerminal = func(int) bool { return true }
	shellMakeRaw = func(int) (*term.State, error) { return nil, errors.New("raw mode unavailable") }

	if _, err := inW.WriteString("exit\n"); err != nil {
		t.Fatalf("write fallback input: %v", err)
	}
	if err := inW.Close(); err != nil {
		t.Fatalf("close fallback input writer: %v", err)
	}

	if err := runCmd("shell"); err != nil {
		t.Fatalf("shell command failed during fallback path: %v", err)
	}
}

func TestShellDefaultHooksCoverage(t *testing.T) {
	setShellHooksForTest(t)

	if shellStdout() == nil {
		t.Fatal("shellStdout should return a file")
	}

	var rw bytes.Buffer
	if shellNewEditor(&rw, "sim-cli> ") == nil {
		t.Fatal("shellNewEditor should return a line editor")
	}
}
