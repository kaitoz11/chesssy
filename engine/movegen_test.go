package engine

import (
	"errors"
	"testing"
)

func TestPinnedPiecesCannotMoveOffTheLine(t *testing.T) {
	// The knight on e2 is pinned by the rook on e8.
	p, err := NewPositionFromFEN("4r2k/8/8/8/8/8/4N3/4K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	var list MoveList
	p.GenerateMoves(&list)
	for _, m := range list.Moves() {
		if m.From() == E2 {
			t.Errorf("pinned knight move generated: %s", m)
		}
	}
}

func TestCheckEvasions(t *testing.T) {
	// Black king on e8 is checked by the rook on e1: block, capture or move.
	p, err := NewPositionFromFEN("4k3/8/8/8/8/8/4R3/4K2R b - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	if !p.InCheck() {
		t.Fatal("expected black to be in check")
	}
	var list MoveList
	p.GenerateMoves(&list)
	got := map[string]bool{}
	for _, m := range list.Moves() {
		got[m.String()] = true
	}
	for _, want := range []string{"e8d8", "e8f8", "e8d7", "e8f7"} {
		if !got[want] {
			t.Errorf("missing evasion %s", want)
		}
	}
	if got["e8e7"] {
		t.Error("e8e7 stays on the checking file and must not be generated")
	}
}

func TestDoubleCheckAllowsOnlyKingMoves(t *testing.T) {
	// Rook on e1 and bishop on h5 both check the black king on e8.
	p, err := NewPositionFromFEN("4k3/6r1/8/7B/8/8/8/4RK2 b - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	if p.Checkers().Count() != 2 {
		t.Fatalf("expected a double check, got %d checkers", p.Checkers().Count())
	}
	var list MoveList
	p.GenerateMoves(&list)
	for _, m := range list.Moves() {
		if m.From() != E8 {
			t.Errorf("non-king move %s generated under double check", m)
		}
	}
}

func TestCheckmateAndStalemate(t *testing.T) {
	tests := []struct {
		name      string
		fen       string
		checkmate bool
		stalemate bool
	}{
		{"back rank mate", "6k1/5ppp/8/8/8/8/8/4R1K1 b - - 0 1", false, false},
		{"mated", "R5k1/5ppp/8/8/8/8/8/6K1 b - - 0 1", true, false},
		{"stalemate", "7k/5Q2/6K1/8/8/8/8/8 b - - 0 1", false, true},
		{"start", StartFEN, false, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := NewPositionFromFEN(tc.fen)
			if err != nil {
				t.Fatal(err)
			}
			if got := p.IsCheckmate(); got != tc.checkmate {
				t.Errorf("IsCheckmate() = %v, want %v", got, tc.checkmate)
			}
			if got := p.IsStalemate(); got != tc.stalemate {
				t.Errorf("IsStalemate() = %v, want %v", got, tc.stalemate)
			}
		})
	}
}

func TestParseMoveRejectsIllegalMoves(t *testing.T) {
	p := NewPosition()
	for _, move := range []string{"e2e5", "e7e5", "", "e2", "e2e4q", "zz99", "e2e4k"} {
		_, err := p.ParseMove(move)
		if err == nil {
			t.Errorf("ParseMove(%q) succeeded, want an error", move)
			continue
		}
		// Malformed squares report ErrInvalidSquare, everything else
		// ErrIllegalMove; both mean "not a move you can play here".
		if !errors.Is(err, ErrIllegalMove) && !errors.Is(err, ErrInvalidSquare) {
			t.Errorf("ParseMove(%q) error %v wraps neither sentinel", move, err)
		}
	}
	if _, err := p.ParseMove("e2e4"); err != nil {
		t.Errorf("ParseMove(\"e2e4\") failed: %v", err)
	}
}

func BenchmarkMoveGeneration(b *testing.B) {
	p, err := NewPositionFromFEN("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1")
	if err != nil {
		b.Fatal(err)
	}
	var list MoveList
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.GenerateMoves(&list)
	}
}
