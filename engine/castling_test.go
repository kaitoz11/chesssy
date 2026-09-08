// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package engine

import (
	"testing"
)

func TestCastlingMoves(t *testing.T) {
	p, err := NewPositionFromFEN("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")
	if err != nil {
		t.Fatal(err)
	}

	m, err := p.ParseMove("e1g1")
	if err != nil {
		t.Fatalf("kingside castling not generated: %v", err)
	}
	p.MakeMove(m)
	if got, want := p.FEN(), "r3k2r/8/8/8/8/8/8/R4RK1 b kq - 1 1"; got != want {
		t.Errorf("after e1g1:\ngot  %q\nwant %q", got, want)
	}
	p.UnmakeMove()

	m, err = p.ParseMove("e1c1")
	if err != nil {
		t.Fatalf("queenside castling not generated: %v", err)
	}
	p.MakeMove(m)
	if got, want := p.FEN(), "r3k2r/8/8/8/8/8/8/2KR3R b kq - 1 1"; got != want {
		t.Errorf("after e1c1:\ngot  %q\nwant %q", got, want)
	}
	p.UnmakeMove()

	// Moving the rook forfeits that side's right only.
	m, _ = p.ParseMove("h1h2")
	p.MakeMove(m)
	if got, want := p.CastlingRights(), WhiteQueenSide|BlackKingSide|BlackQueenSide; got != want {
		t.Errorf("castling rights after h1h2 = %v, want %v", got, want)
	}
	p.UnmakeMove()
	if got := p.CastlingRights(); got != AllCastling {
		t.Errorf("castling rights after unmake = %v, want %v", got, AllCastling)
	}
}

func TestCastlingIsIllegalThroughAttackedSquares(t *testing.T) {
	tests := []struct {
		name  string
		fen   string
		move  string
		legal bool
	}{
		{"clear path", "4k3/8/8/8/8/8/8/4K2R w K - 0 1", "e1g1", true},
		{"rook attacks f1", "4k3/8/8/8/8/8/5q2/4K2R w K - 0 1", "e1g1", false},
		{"rook attacks g1", "4k3/8/8/8/8/8/6q1/4K2R w K - 0 1", "e1g1", false},
		{"king in check", "4k3/8/8/8/8/8/4q3/4K2R w K - 0 1", "e1g1", false},
		{"blocked path", "4k3/8/8/8/8/8/8/4KN1R w K - 0 1", "e1g1", false},
		// b1 is not on the king's path, so an attack there does not matter.
		{"queenside b1 attacked is legal", "4k3/8/8/8/8/8/3n4/R3K3 w Q - 0 1", "e1c1", true},
		{"queenside d1 attacked", "4k3/8/8/8/8/8/3q4/R3K3 w Q - 0 1", "e1c1", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := NewPositionFromFEN(tc.fen)
			if err != nil {
				t.Fatal(err)
			}
			_, err = p.ParseMove(tc.move)
			if legal := err == nil; legal != tc.legal {
				t.Errorf("%s legal = %v, want %v", tc.move, legal, tc.legal)
			}
		})
	}
}
