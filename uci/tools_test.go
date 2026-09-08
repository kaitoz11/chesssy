package uci

import (
	"strings"
	"testing"
)

func TestPerftCommand(t *testing.T) {
	got := run(t, "position fen r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", "perft 3", "quit")
	if !strings.Contains(got, "nodes: 97862") {
		t.Errorf("expected kiwipete perft(3) = 97862:\n%s", got)
	}
}

func TestBenchCommand(t *testing.T) {
	got := run(t, "bench 4", "quit")
	if !strings.Contains(got, "bench: depth 4") {
		t.Errorf("bench summary missing:\n%s", got)
	}
	if !strings.Contains(got, "nodes/s") {
		t.Errorf("bench did not report a speed:\n%s", got)
	}
}
