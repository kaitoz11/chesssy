// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package engine

// Zobrist hashing gives each position a 64-bit key that can be updated
// incrementally: toggling a piece on a square, the castling rights, the en
// passant file or the side to move is one exclusive or each. The transposition
// table and repetition detection both key off it.
//
// The keys come from a fixed seed, so hashes are stable between runs. That keeps
// test expectations and any stored data reproducible.

const zobristSeed = 0x1D2B_9A73_6E45_C081

var (
	zobristPieces    [PieceCount][SquareCount]uint64
	zobristCastling  [AllCastling + 1]uint64
	zobristEnPassant [FileCount]uint64
	zobristSide      uint64
)

func init() {
	rng := newRandom(zobristSeed)
	for piece := WhitePawn; piece < NoPiece; piece++ {
		for s := A1; s < SquareCount; s++ {
			zobristPieces[piece][s] = rng.uint64()
		}
	}
	for i := range zobristCastling {
		zobristCastling[i] = rng.uint64()
	}
	for i := range zobristEnPassant {
		zobristEnPassant[i] = rng.uint64()
	}
	zobristSide = rng.uint64()
}
