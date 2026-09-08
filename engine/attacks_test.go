// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package engine

import (
	"testing"
)

func TestLeaperAttacks(t *testing.T) {
	tests := []struct {
		name string
		got  Bitboard
		want []Square
	}{
		{"knight corner", KnightAttacks(A1), []Square{B3, C2}},
		{"knight center", KnightAttacks(D4), []Square{C2, E2, B3, F3, B5, F5, C6, E6}},
		{"king corner", KingAttacks(H8), []Square{G8, G7, H7}},
		{"king center", KingAttacks(E4), []Square{D3, E3, F3, D4, F4, D5, E5, F5}},
		{"white pawn", PawnAttacks(White, D4), []Square{C5, E5}},
		{"black pawn", PawnAttacks(Black, D4), []Square{C3, E3}},
		{"white pawn edge", PawnAttacks(White, A2), []Square{B3}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var want Bitboard
			for _, s := range tc.want {
				want = want.With(s)
			}
			if tc.got != want {
				t.Errorf("got:\n%v\nwant:\n%v", tc.got, want)
			}
		})
	}
}

func TestBetweenAndLine(t *testing.T) {
	if got, want := BetweenBB(A1, A4), SquareBB(A2)|SquareBB(A3); got != want {
		t.Errorf("BetweenBB(A1, A4):\ngot:\n%v\nwant:\n%v", got, want)
	}
	if got := BetweenBB(A1, B3); got != EmptyBB {
		t.Errorf("BetweenBB(A1, B3) = %v, want empty (not aligned)", got)
	}
	if got, want := BetweenBB(C1, F4), SquareBB(D2)|SquareBB(E3); got != want {
		t.Errorf("BetweenBB(C1, F4):\ngot:\n%v\nwant:\n%v", got, want)
	}
	if got, want := LineBB(A1, A4), FileBB(FileA); got != want {
		t.Errorf("LineBB(A1, A4):\ngot:\n%v\nwant:\n%v", got, want)
	}
	if got, want := LineBB(B2, C2), RankBB(Rank2); got != want {
		t.Errorf("LineBB(B2, C2):\ngot:\n%v\nwant:\n%v", got, want)
	}
	if !Aligned(A1, D4, H8) {
		t.Error("Aligned(A1, D4, H8) = false, want true")
	}
	if Aligned(A1, D4, H7) {
		t.Error("Aligned(A1, D4, H7) = true, want false")
	}
}
