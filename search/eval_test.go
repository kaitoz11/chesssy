package search

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kaitoz11/chesssy/engine"
)

// mirrorFEN swaps colors and flips the board vertically. A correct evaluation
// must score a position and its mirror identically, since both sides are
// evaluated with the same terms and the score is relative to the side to move.
func mirrorFEN(t *testing.T, fen string) string {
	t.Helper()
	fields := strings.Fields(fen)
	if len(fields) < 4 {
		t.Fatalf("mirrorFEN: bad FEN %q", fen)
	}

	ranks := strings.Split(fields[0], "/")
	for i, j := 0, len(ranks)-1; i < j; i, j = i+1, j-1 {
		ranks[i], ranks[j] = ranks[j], ranks[i]
	}
	for i, rank := range ranks {
		ranks[i] = strings.Map(swapCase, rank)
	}
	fields[0] = strings.Join(ranks, "/")

	if fields[1] == "w" {
		fields[1] = "b"
	} else {
		fields[1] = "w"
	}
	fields[2] = strings.Map(swapCase, fields[2])
	if fields[3] != "-" {
		fields[3] = string(fields[3][0]) + string('1'+('8'-fields[3][1]))
	}
	return strings.Join(fields, " ")
}

func swapCase(r rune) rune {
	switch {
	case r >= 'a' && r <= 'z':
		return r - 32
	case r >= 'A' && r <= 'Z':
		return r + 32
	default:
		return r
	}
}

func TestEvaluateIsColorSymmetric(t *testing.T) {
	fens := []string{
		engine.StartFEN,
		"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		"r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
		"4rrk1/pp1n3p/3q2pQ/2p1pb2/2PP4/2P3N1/P2B2PP/4RRK1 b - - 7 19",
		"6k1/6p1/6Pp/ppp5/3pn2P/1P3K2/1PP2P2/3N4 b - - 0 1",
		"8/8/8/2k5/2pP4/8/B7/4K3 b - d3 0 3",
	}
	for _, fen := range fens {
		t.Run(fen, func(t *testing.T) {
			p, err := engine.NewPositionFromFEN(fen)
			if err != nil {
				t.Fatal(err)
			}
			mirrored := mirrorFEN(t, fen)
			q, err := engine.NewPositionFromFEN(mirrored)
			if err != nil {
				t.Fatalf("mirrored FEN %q: %v", mirrored, err)
			}
			if got, want := Evaluate(q), Evaluate(p); got != want {
				t.Errorf("mirror of %q scored %d, original scored %d", fen, got, want)
			}
		})
	}
}

func TestEvaluateMaterialAdvantage(t *testing.T) {
	tests := []struct {
		name    string
		fen     string
		atLeast int
	}{
		{"extra queen", "4k3/8/8/8/8/8/8/3QK3 w - - 0 1", 700},
		{"extra rook", "4k3/8/8/8/8/8/8/3RK3 w - - 0 1", 350},
		{"extra pawn", "4k3/8/8/8/8/8/3P4/4K3 w - - 0 1", 50},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := engine.NewPositionFromFEN(tc.fen)
			if err != nil {
				t.Fatal(err)
			}
			if got := Evaluate(p); got < tc.atLeast {
				t.Errorf("Evaluate() = %d, want at least %d", got, tc.atLeast)
			}
		})
	}
}

func TestEvaluateBalancedPositionIsNearZero(t *testing.T) {
	p := engine.NewPosition()
	if got := Evaluate(p); got < -60 || got > 60 {
		t.Errorf("start position evaluated at %d, want roughly balanced", got)
	}
}

func TestEvaluatePassedPawnsAndStructure(t *testing.T) {
	// A passed pawn on the sixth is worth more than a blocked one on the second.
	advanced, err := engine.NewPositionFromFEN("4k3/8/2P5/8/8/8/8/4K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	back, err := engine.NewPositionFromFEN("4k3/8/8/8/8/8/2P5/4K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	if Evaluate(advanced) <= Evaluate(back) {
		t.Errorf("advanced passer scored %d, back pawn scored %d", Evaluate(advanced), Evaluate(back))
	}

	// Doubled pawns are worse than pawns on adjacent files.
	doubled, err := engine.NewPositionFromFEN("4k3/8/8/8/8/2P5/2P5/4K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	spread, err := engine.NewPositionFromFEN("4k3/8/8/8/8/2P5/3P4/4K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	if Evaluate(doubled) >= Evaluate(spread) {
		t.Errorf("doubled pawns scored %d, spread pawns scored %d", Evaluate(doubled), Evaluate(spread))
	}
}

func TestSeeValuesOrdering(t *testing.T) {
	// Move ordering relies on the exchange values being monotonic.
	values := []int{
		engine.SeeValue(engine.Pawn),
		engine.SeeValue(engine.Knight),
		engine.SeeValue(engine.Rook),
		engine.SeeValue(engine.Queen),
	}
	for i := 1; i < len(values); i++ {
		if values[i] <= values[i-1] {
			t.Errorf("SeeValues not increasing: %v", values)
		}
	}
}

func BenchmarkEvaluate(b *testing.B) {
	p, err := engine.NewPositionFromFEN("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1")
	if err != nil {
		b.Fatal(err)
	}
	var sink int
	for i := 0; i < b.N; i++ {
		sink += Evaluate(p)
	}
	if sink == 0 {
		b.Fatal("unreachable")
	}
}

func BenchmarkSearch(b *testing.B) {
	for _, depth := range []int{8, 10} {
		b.Run(fmt.Sprintf("depth-%d", depth), func(b *testing.B) {
			p := engine.NewPosition()
			s := NewSearcher(64)
			var nodes uint64
			for i := 0; i < b.N; i++ {
				s.NewGame()
				info := s.Search(p, Limits{Depth: depth})
				nodes += info.Nodes
			}
			b.ReportMetric(float64(nodes)/b.Elapsed().Seconds(), "nodes/s")
		})
	}
}
