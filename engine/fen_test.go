package engine

import (
	"errors"
	"strings"
	"testing"
)

func TestFENRoundTrip(t *testing.T) {
	fens := []string{
		StartFEN,
		"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		"rnbqkb1r/ppp1pppp/3p1n2/8/3PP3/2N5/PPP2PPP/R1BQKBNR b KQkq d3 0 3",
		"4k3/8/8/8/8/8/8/4K3 w - - 42 99",
		"8/8/8/2k5/2pP4/8/B7/4K3 b - d3 0 3",
	}
	for _, fen := range fens {
		t.Run(fen, func(t *testing.T) {
			p, err := NewPositionFromFEN(fen)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if got := p.FEN(); got != fen {
				t.Errorf("FEN() = %q, want %q", got, fen)
			}
		})
	}
}

func TestFENErrors(t *testing.T) {
	tests := []struct {
		name string
		fen  string
	}{
		{"too few fields", "8/8/8/8/8/8/8/8 w"},
		{"too few ranks", "8/8/8/8/8/8/8 w - - 0 1"},
		{"rank too short", "7/8/8/8/8/8/8/8 w - - 0 1"},
		{"bad piece", "xnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
		{"bad side", "8/8/8/8/8/8/8/8 x - - 0 1"},
		{"bad castling", "4k3/8/8/8/8/8/8/4K3 w Xq - 0 1"},
		{"bad en passant", "4k3/8/8/8/8/8/8/4K3 w - z9 0 1"},
		{"bad halfmove clock", "4k3/8/8/8/8/8/8/4K3 w - - x 1"},
		{"bad fullmove number", "4k3/8/8/8/8/8/8/4K3 w - - 0 0"},
		{"no kings", "8/8/8/8/8/8/8/8 w - - 0 1"},
		{"two white kings", "4k3/8/8/8/8/8/8/3KK3 w - - 0 1"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewPositionFromFEN(tc.fen)
			if err == nil {
				t.Fatal("expected an error")
			}
			// Callers classify parse failures with errors.Is, so every one of
			// them has to carry the sentinel.
			if !errors.Is(err, ErrInvalidFEN) {
				t.Errorf("error %v does not wrap ErrInvalidFEN", err)
			}
		})
	}
}

func TestSetFENLeavesPositionUnchangedOnError(t *testing.T) {
	p := NewPosition()
	before := p.FEN()
	if err := p.SetFEN("garbage"); err == nil {
		t.Fatal("expected an error")
	}
	if got := p.FEN(); got != before {
		t.Errorf("position changed to %q after a failed SetFEN", got)
	}
}

func TestPositionStringRendersBoard(t *testing.T) {
	got := NewPosition().String()
	for _, want := range []string{"8: r n b q k b n r", "1: R N B Q K B N R", "a b c d e f g h", "w KQkq - 0 1"} {
		if !strings.Contains(got, want) {
			t.Errorf("board rendering missing %q:\n%s", want, got)
		}
	}
}
