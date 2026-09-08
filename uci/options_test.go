package uci

import (
	"strings"
	"testing"
)

func TestSetOption(t *testing.T) {
	got := run(t, "setoption name Hash value 4", "setoption name Threads value 4",
		"setoption name Clear Hash", "setoption name Nonsense value 1", "isready", "quit")
	if !strings.Contains(got, "readyok") {
		t.Errorf("engine stopped responding after setoption:\n%s", got)
	}
	if !strings.Contains(got, `unknown option "Nonsense"`) {
		t.Errorf("unknown option not reported:\n%s", got)
	}
	if strings.Contains(got, `unknown option "Hash"`) || strings.Contains(got, `unknown option "Clear Hash"`) {
		t.Errorf("known options reported as unknown:\n%s", got)
	}
}
