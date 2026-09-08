package uci

import (
	"bytes"
	"strings"
	"testing"
)

// run executes commands against a fresh engine and returns everything it wrote.
// Searches started by "go" are waited for, so the output is complete and the
// test does not race the search goroutine.
func run(t *testing.T, commands ...string) string {
	t.Helper()
	var out bytes.Buffer
	e := New(&out)
	for _, c := range commands {
		// A GUI waits for bestmove before quitting; do the same so that tests
		// observe complete searches instead of aborted ones.
		if c == "quit" {
			e.searches.Wait()
		}
		if quit := e.Execute(c); quit {
			break
		}
	}
	e.searches.Wait()
	return out.String()
}

// TestRunReadsCommandsFromReader covers the scanner loop, which the other tests
// bypass by calling Execute directly.

func TestRunReadsCommandsFromReader(t *testing.T) {
	var out bytes.Buffer
	e := New(&out)
	if err := e.Run(strings.NewReader("uci\nisready\nposition startpos\nfen\nquit\n")); err != nil {
		t.Fatalf("Run: %v", err)
	}
	got := out.String()
	for _, want := range []string{"uciok", "readyok", "rnbqkbnr"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestHandshake(t *testing.T) {
	got := run(t, "uci", "isready", "quit")
	for _, want := range []string{"id name " + Name, "id author " + Author, "option name Hash", "uciok", "readyok"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestUciNewGameResetsPosition(t *testing.T) {
	got := run(t, "position startpos moves e2e4", "ucinewgame", "fen", "quit")
	if !strings.Contains(got, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1") {
		t.Errorf("ucinewgame did not reset the board:\n%s", got)
	}
}

func TestDebugCommands(t *testing.T) {
	got := run(t, "d", "eval", "help", "nonsense", "quit")
	if !strings.Contains(got, "a b c d e f g h") {
		t.Errorf("d did not print a board:\n%s", got)
	}
	if !strings.Contains(got, "static eval") {
		t.Errorf("eval did not print a score:\n%s", got)
	}
	if !strings.Contains(got, "commands:") {
		t.Errorf("help did not print usage:\n%s", got)
	}
	if !strings.Contains(got, `unknown command "nonsense"`) {
		t.Errorf("unknown command not reported:\n%s", got)
	}
}

func TestEmptyAndUnknownInputIsSafe(t *testing.T) {
	got := run(t, "", "   ", "quit")
	if strings.Contains(got, "panic") {
		t.Errorf("blank input broke the engine:\n%s", got)
	}
}
