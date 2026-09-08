package engine

import (
	"testing"
)

func TestSeeGE(t *testing.T) {
	tests := []struct {
		name      string
		fen       string
		move      string
		threshold int
		want      bool
	}{
		{
			name: "free pawn", fen: "4k3/8/8/3p4/4P3/8/8/4K3 w - - 0 1",
			move: "e4d5", threshold: 0, want: true,
		},
		{
			name: "defended pawn costs a knight", fen: "4k3/2p5/3p4/8/4N3/8/8/4K3 w - - 0 1",
			move: "e4d6", threshold: 0, want: false,
		},
		{
			name: "equal trade meets threshold zero", fen: "4k3/8/2n5/3p4/4N3/8/8/4K3 w - - 0 1",
			move: "e4d6", threshold: 0, want: true,
		},
		{
			name: "rook takes defended rook", fen: "4k3/8/8/8/8/4r3/4r3/4RK2 w - - 0 1",
			move: "e1e2", threshold: 0, want: true,
		},
		{
			name: "queen takes defended pawn", fen: "4k3/1p6/2p5/8/8/8/8/K5Q1 w - - 0 1",
			move: "g1c5", threshold: 0, want: true,
		},
		{
			name: "capture worth at least a rook", fen: "4k3/8/8/3r4/4P3/8/8/4K3 w - - 0 1",
			move: "e4d5", threshold: 400, want: true,
		},
		{
			name: "capture not worth a queen", fen: "4k3/8/8/3r4/4P3/8/8/4K3 w - - 0 1",
			move: "e4d5", threshold: 900, want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := NewPositionFromFEN(tc.fen)
			if err != nil {
				t.Fatal(err)
			}
			m, err := p.ParseMove(tc.move)
			if err != nil {
				t.Fatal(err)
			}
			if got := p.SeeGE(m, tc.threshold); got != tc.want {
				t.Errorf("SeeGE(%s, %d) = %v, want %v (see = %d)", tc.move, tc.threshold, got, tc.want, p.See(m))
			}
		})
	}
}

func TestSeeValuesFreeCapture(t *testing.T) {
	p, err := NewPositionFromFEN("4k3/8/8/3q4/4P3/8/8/4K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	m, err := p.ParseMove("e4d5")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := p.See(m), SeeValue(Queen); got != want {
		t.Errorf("See(e4d5) = %d, want %d", got, want)
	}
}
