// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package engine

import (
	"fmt"
	"testing"
)

func TestEnPassant(t *testing.T) {
	p, err := NewPositionFromFEN("4k3/8/8/8/4p3/8/3P4/4K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	m, err := p.ParseMove("d2d4")
	if err != nil {
		t.Fatal(err)
	}
	p.MakeMove(m)
	if got := p.EnPassantSquare(); got != D3 {
		t.Fatalf("en passant square = %v, want d3", got)
	}

	capture, err := p.ParseMove("e4d3")
	if err != nil {
		t.Fatalf("en passant capture not generated: %v", err)
	}
	if capture.Kind() != EnPassantMove {
		t.Fatalf("e4d3 kind = %v, want EnPassantMove", capture.Kind())
	}
	before := p.FEN()
	p.MakeMove(capture)
	if got, want := p.FEN(), "4k3/8/8/8/8/3p4/8/4K3 w - - 0 2"; got != want {
		t.Errorf("after e4d3:\ngot  %q\nwant %q", got, want)
	}
	p.UnmakeMove()
	if got := p.FEN(); got != before {
		t.Errorf("unmake en passant:\ngot  %q\nwant %q", got, before)
	}
}

func TestEnPassantIsIllegalWhenItExposesTheKing(t *testing.T) {
	// White rook on h5 pins both pawns to the black king on a5: capturing en
	// passant removes two pieces from the rank and exposes the king.
	p, err := NewPositionFromFEN("8/8/8/k1pP3R/8/8/8/4K3 b - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.ParseMove("c5c4"); err != nil {
		t.Fatalf("c5c4 should be legal: %v", err)
	}

	p, err = NewPositionFromFEN("8/8/8/8/k1pP3R/8/8/4K3 b - d3 0 1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.ParseMove("c4d3"); err == nil {
		t.Error("en passant capture should be illegal: it exposes the king on the rank")
	}
}

func TestPromotions(t *testing.T) {
	p, err := NewPositionFromFEN("4k2r/6P1/8/8/8/8/8/4K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	var list MoveList
	p.GenerateMoves(&list)

	promotions := map[string]bool{}
	for _, m := range list.Moves() {
		if m.IsPromotion() {
			promotions[m.String()] = true
		}
	}
	// Both the quiet push to g8 and the capture on h8 promote to four pieces.
	for _, want := range []string{"g7g8q", "g7g8r", "g7g8b", "g7g8n", "g7h8q", "g7h8r", "g7h8b", "g7h8n"} {
		if !promotions[want] {
			t.Errorf("missing promotion %s", want)
		}
	}
	if len(promotions) != 8 {
		t.Errorf("got %d promotions, want 8: %v", len(promotions), promotions)
	}

	m, _ := p.ParseMove("g7h8q")
	p.MakeMove(m)
	if got, want := p.FEN(), "4k2Q/8/8/8/8/8/8/4K3 b - - 0 1"; got != want {
		t.Errorf("after g7h8q:\ngot  %q\nwant %q", got, want)
	}
	p.UnmakeMove()
	if got := p.FEN(); got != "4k2r/6P1/8/8/8/8/8/4K3 w - - 0 1" {
		t.Errorf("unmake promotion capture gave %q", p.FEN())
	}
}

func TestNullMoveIsReversible(t *testing.T) {
	p, err := NewPositionFromFEN("rnbqkb1r/ppp1pppp/3p1n2/8/3PP3/2N5/PPP2PPP/R1BQKBNR b KQkq d3 0 3")
	if err != nil {
		t.Fatal(err)
	}
	before, beforeKey := p.FEN(), p.Key()
	p.MakeNullMove()
	if p.SideToMove() != White {
		t.Error("null move did not pass the turn")
	}
	if p.EnPassantSquare() != NoSquare {
		t.Error("null move must clear the en passant square")
	}
	p.UnmakeNullMove()
	if got := p.FEN(); got != before {
		t.Errorf("FEN after null move round trip = %q, want %q", got, before)
	}
	if p.Key() != beforeKey {
		t.Errorf("key after null move round trip = %#x, want %#x", p.Key(), beforeKey)
	}
}

func TestZobristKeyIsPathIndependent(t *testing.T) {
	// The same position reached by different move orders must hash the same.
	// Only knight moves are used, so the halfmove clock and en passant square
	// match as well.
	a := NewPosition()
	for _, move := range []string{"g1f3", "g8f6", "b1c3", "b8c6"} {
		m, _ := a.ParseMove(move)
		a.MakeMove(m)
	}
	b := NewPosition()
	for _, move := range []string{"b1c3", "b8c6", "g1f3", "g8f6"} {
		m, _ := b.ParseMove(move)
		b.MakeMove(m)
	}
	if a.FEN() != b.FEN() {
		t.Fatalf("positions differ:\n%s\n%s", a.FEN(), b.FEN())
	}
	if a.Key() != b.Key() {
		t.Errorf("keys differ: %#x vs %#x", a.Key(), b.Key())
	}
}

// TestMakeUnmakeRestoresPosition walks the tree and verifies that unmaking a
// move restores the FEN and the incrementally updated Zobrist key.
func TestMakeUnmakeRestoresPosition(t *testing.T) {
	for _, tc := range perftSuite {
		t.Run(tc.name, func(t *testing.T) {
			p, err := NewPositionFromFEN(tc.fen)
			if err != nil {
				t.Fatalf("parse FEN: %v", err)
			}
			if err := walkAndVerify(p, 3); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func walkAndVerify(p *Position, depth int) error {
	if depth == 0 {
		return nil
	}
	before, beforeKey := p.FEN(), p.key

	var list MoveList
	p.GenerateMoves(&list)
	for _, m := range list.Moves() {
		p.MakeMove(m)
		if got, want := p.key, p.computeKey(); got != want {
			return fmt.Errorf("after %s from %s: incremental key %#x, recomputed %#x", m, before, got, want)
		}
		if err := walkAndVerify(p, depth-1); err != nil {
			return err
		}
		p.UnmakeMove()
		if got := p.FEN(); got != before {
			return fmt.Errorf("unmake %s: FEN %s, want %s", m, got, before)
		}
		if p.key != beforeKey {
			return fmt.Errorf("unmake %s: key %#x, want %#x", m, p.key, beforeKey)
		}
	}
	return nil
}
