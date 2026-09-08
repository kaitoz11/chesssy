package uci

import (
	"strings"
	"testing"
)

func TestGoDepthProducesBestMove(t *testing.T) {
	got := run(t, "position startpos", "go depth 6", "quit")
	if !strings.Contains(got, "info depth 6") {
		t.Errorf("missing depth 6 info line:\n%s", got)
	}
	if !strings.Contains(got, "bestmove ") {
		t.Errorf("missing bestmove:\n%s", got)
	}
	for _, field := range []string{"seldepth", "score cp", "nodes", "nps", "hashfull", "time", "pv"} {
		if !strings.Contains(got, field) {
			t.Errorf("info line missing %q:\n%s", field, got)
		}
	}
}

func TestGoReportsMate(t *testing.T) {
	got := run(t, "position fen 6k1/5ppp/8/8/8/8/8/R5K1 w - - 0 1", "go depth 4", "quit")
	if !strings.Contains(got, "score mate 1") {
		t.Errorf("expected a mate score:\n%s", got)
	}
	if !strings.Contains(got, "bestmove a1a8") {
		t.Errorf("expected bestmove a1a8:\n%s", got)
	}
}

func TestGoOnMatedPositionStillReplies(t *testing.T) {
	got := run(t, "position fen R5k1/5ppp/8/8/8/8/8/6K1 b - - 0 1", "go depth 4", "quit")
	if !strings.Contains(got, "bestmove") {
		t.Errorf("engine must always answer with bestmove:\n%s", got)
	}
}

func TestGoMoveTimeAndStop(t *testing.T) {
	got := run(t, "position startpos", "go movetime 100", "quit")
	if !strings.Contains(got, "bestmove") {
		t.Errorf("movetime search produced no bestmove:\n%s", got)
	}

	got = run(t, "position startpos", "go infinite", "stop", "quit")
	if !strings.Contains(got, "bestmove") {
		t.Errorf("stopped search produced no bestmove:\n%s", got)
	}
}

func TestGoWithClock(t *testing.T) {
	got := run(t, "position startpos", "go wtime 400 btime 400 winc 10 binc 10", "quit")
	if !strings.Contains(got, "bestmove") {
		t.Errorf("clock search produced no bestmove:\n%s", got)
	}
}

func TestGoPerft(t *testing.T) {
	got := run(t, "position startpos", "go perft 3", "quit")
	if !strings.Contains(got, "nodes: 8902") {
		t.Errorf("expected perft(3) = 8902:\n%s", got)
	}
	if !strings.Contains(got, "a2a3: 380") {
		t.Errorf("expected divide output per move:\n%s", got)
	}
}
