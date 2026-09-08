// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package engine

import "strings"

// CastlingRights is a bit set of the four castling privileges.
type CastlingRights uint8

const (
	WhiteKingSide CastlingRights = 1 << iota
	WhiteQueenSide
	BlackKingSide
	BlackQueenSide

	NoCastling  CastlingRights = 0
	AllCastling                = WhiteKingSide | WhiteQueenSide | BlackKingSide | BlackQueenSide
)

// String returns the rights in FEN form, such as "Kq", or "-" for none.
func (cr CastlingRights) String() string {
	if cr == NoCastling {
		return "-"
	}
	var sb strings.Builder
	for i, c := range []byte("KQkq") {
		if cr&(1<<uint(i)) != 0 {
			sb.WriteByte(c)
		}
	}
	return sb.String()
}

// castlingInfo records the squares one castling privilege involves.
type castlingInfo struct {
	kingFrom, kingTo Square
	rookFrom, rookTo Square
	emptyPath        Bitboard // must be unoccupied
	safePath         Bitboard // must not be attacked; the king's origin is checked separately
}

// castlings is indexed by the bit position of the corresponding CastlingRights
// value: 0 = white kingside, 1 = white queenside, 2 = black kingside,
// 3 = black queenside. castlingIndexOf maps a colour and king destination back
// to that index.
var castlings = [4]castlingInfo{
	{E1, G1, H1, F1, SquareBB(F1) | SquareBB(G1), SquareBB(F1) | SquareBB(G1)},
	{E1, C1, A1, D1, SquareBB(B1) | SquareBB(C1) | SquareBB(D1), SquareBB(C1) | SquareBB(D1)},
	{E8, G8, H8, F8, SquareBB(F8) | SquareBB(G8), SquareBB(F8) | SquareBB(G8)},
	{E8, C8, A8, D8, SquareBB(B8) | SquareBB(C8) | SquareBB(D8), SquareBB(C8) | SquareBB(D8)},
}

// castlingIndexOf returns the castlings index for a castling move by c whose
// king lands on kingTo.
func castlingIndexOf(c Color, kingTo Square) int {
	side := 0 // kingside
	if kingTo.File() == FileC {
		side = 1
	}
	return side + 2*int(c)
}

// castlingLoss[s] holds the rights forfeited when a piece moves from or to
// square s, which covers king moves as well as rook moves and rook captures.
var castlingLoss [SquareCount]CastlingRights

func init() {
	castlingLoss[E1] = WhiteKingSide | WhiteQueenSide
	castlingLoss[H1] = WhiteKingSide
	castlingLoss[A1] = WhiteQueenSide
	castlingLoss[E8] = BlackKingSide | BlackQueenSide
	castlingLoss[H8] = BlackKingSide
	castlingLoss[A8] = BlackQueenSide
}
