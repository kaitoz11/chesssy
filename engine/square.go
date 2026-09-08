// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package engine

import "fmt"

// File is a board column, A through H.
type File uint8

const (
	FileA File = iota
	FileB
	FileC
	FileD
	FileE
	FileF
	FileG
	FileH

	// FileCount is the number of files, handy as an array size.
	FileCount = 8
)

// Rank is a board row, 1 through 8.
type Rank uint8

const (
	Rank1 Rank = iota
	Rank2
	Rank3
	Rank4
	Rank5
	Rank6
	Rank7
	Rank8

	// RankCount is the number of ranks, handy as an array size.
	RankCount = 8
)

// Square identifies one of the 64 board squares, counted from A1 = 0 along each
// rank. NoSquare (64) is the sentinel for "no square", such as a position
// without an en passant target.
type Square uint8

const (
	A1 Square = iota
	B1
	C1
	D1
	E1
	F1
	G1
	H1
	A2
	B2
	C2
	D2
	E2
	F2
	G2
	H2
	A3
	B3
	C3
	D3
	E3
	F3
	G3
	H3
	A4
	B4
	C4
	D4
	E4
	F4
	G4
	H4
	A5
	B5
	C5
	D5
	E5
	F5
	G5
	H5
	A6
	B6
	C6
	D6
	E6
	F6
	G6
	H6
	A7
	B7
	C7
	D7
	E7
	F7
	G7
	H7
	A8
	B8
	C8
	D8
	E8
	F8
	G8
	H8

	NoSquare
	// SquareCount is the number of squares, handy as an array size.
	SquareCount = 64
)

// NewSquare builds a square from a file and a rank.
func NewSquare(f File, r Rank) Square { return Square(uint8(r)*8 + uint8(f)) }

// File returns the square's column.
func (s Square) File() File { return File(uint8(s) & 7) }

// Rank returns the square's row.
func (s Square) Rank() Rank { return Rank(uint8(s) >> 3) }

// Flip mirrors the square vertically, mapping A1 to A8. It lets white-oriented
// tables be reused for black.
func (s Square) Flip() Square { return s ^ 56 }

// IsValid reports whether s addresses a real board square.
func (s Square) IsValid() bool { return s < SquareCount }

// Shift returns the square d steps away. Staying on the board is the caller's
// responsibility; it is used where that is known, such as pawn pushes.
func (s Square) Shift(d Direction) Square { return Square(int8(s) + int8(d)) }

// String returns algebraic notation such as "e4", or "-" for NoSquare.
func (s Square) String() string {
	if !s.IsValid() {
		return "-"
	}
	return fmt.Sprintf("%c%c", 'a'+byte(s.File()), '1'+byte(s.Rank()))
}

// SquareFromString parses algebraic square notation such as "e4".
func SquareFromString(s string) (Square, error) {
	if len(s) != 2 || s[0] < 'a' || s[0] > 'h' || s[1] < '1' || s[1] > '8' {
		return NoSquare, fmt.Errorf("%w: %q", ErrInvalidSquare, s)
	}
	return NewSquare(File(s[0]-'a'), Rank(s[1]-'1')), nil
}

// Direction is a square offset, used both to step squares and to shift whole
// bitboards.
type Direction int8

const (
	North     Direction = 8
	South     Direction = -8
	East      Direction = 1
	West      Direction = -1
	NorthEast Direction = North + East
	NorthWest Direction = North + West
	SouthEast Direction = South + East
	SouthWest Direction = South + West
)

// PawnPush returns the direction pawns of color c move in.
func PawnPush(c Color) Direction {
	if c == White {
		return North
	}
	return South
}
