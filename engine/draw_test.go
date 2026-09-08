// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package engine

import (
	"testing"
)

func TestRepetitionAndFiftyMoveDraws(t *testing.T) {
	p := NewPosition()
	if p.IsRepetition() {
		t.Error("start position reports a repetition")
	}
	// Shuffle knights back and forth: the position after four moves repeats.
	for _, move := range []string{"g1f3", "g8f6", "f3g1", "f6g8"} {
		m, err := p.ParseMove(move)
		if err != nil {
			t.Fatalf("%s: %v", move, err)
		}
		p.MakeMove(m)
	}
	if !p.IsRepetition() {
		t.Error("expected a repetition after returning to the start position")
	}
	if p.HalfMoveClock() != 4 {
		t.Errorf("halfmove clock = %d, want 4", p.HalfMoveClock())
	}

	p, err := NewPositionFromFEN("4k3/8/8/8/8/8/8/R3K3 w - - 99 60")
	if err != nil {
		t.Fatal(err)
	}
	if p.IsFiftyMoveDraw() {
		t.Error("99 half moves is not yet a draw")
	}
	m, _ := p.ParseMove("a1a2")
	p.MakeMove(m)
	if !p.IsFiftyMoveDraw() {
		t.Error("100 half moves should be a draw")
	}
}

func TestInsufficientMaterial(t *testing.T) {
	tests := []struct {
		fen  string
		want bool
	}{
		{"4k3/8/8/8/8/8/8/4K3 w - - 0 1", true},
		{"4k3/8/8/8/8/8/8/3BK3 w - - 0 1", true},
		{"4k3/8/8/8/8/8/8/3NK3 w - - 0 1", true},
		{"3bk3/8/8/8/8/8/8/3BK3 w - - 0 1", false},
		{"4k3/8/8/8/8/8/4P3/4K3 w - - 0 1", false},
		{"4k3/8/8/8/8/8/8/3RK3 w - - 0 1", false},
	}
	for _, tc := range tests {
		t.Run(tc.fen, func(t *testing.T) {
			p, err := NewPositionFromFEN(tc.fen)
			if err != nil {
				t.Fatal(err)
			}
			if got := p.HasInsufficientMaterial(); got != tc.want {
				t.Errorf("HasInsufficientMaterial() = %v, want %v", got, tc.want)
			}
		})
	}
}
