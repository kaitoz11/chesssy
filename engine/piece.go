// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package engine

import "strings"

// Color is the side of a piece, or the side to move.
type Color uint8

const (
	White Color = iota
	Black

	// ColorCount is the number of colors, handy as an array size.
	ColorCount = 2
)

// Flip returns the opposing color.
func (c Color) Flip() Color { return c ^ 1 }

// String returns "white" or "black".
func (c Color) String() string {
	if c == White {
		return "white"
	}
	return "black"
}

// PieceType is a piece kind without color.
type PieceType uint8

const (
	Pawn PieceType = iota
	Knight
	Bishop
	Rook
	Queen
	King

	// PieceTypeCount is the number of piece kinds, handy as an array size.
	PieceTypeCount = 6
)

// String returns the lower case piece letter, as used by FEN for black pieces
// and by UCI for promotions.
func (pt PieceType) String() string { return string(rune("pnbrqk"[pt])) }

// Piece is a colored piece. White pieces come first, so that
// Piece(color*PieceTypeCount + pieceType) holds.
type Piece uint8

const (
	WhitePawn Piece = iota
	WhiteKnight
	WhiteBishop
	WhiteRook
	WhiteQueen
	WhiteKing
	BlackPawn
	BlackKnight
	BlackBishop
	BlackRook
	BlackQueen
	BlackKing

	// NoPiece marks an empty square. It also bounds the real pieces, so
	// NoPiece doubles as the piece count when sizing arrays.
	NoPiece
	PieceCount = 12
)

// pieceChars lists the FEN letters in Piece order.
const pieceChars = "PNBRQKpnbrqk"

// NewPiece combines a color and a piece type.
func NewPiece(c Color, pt PieceType) Piece { return Piece(uint8(c)*PieceTypeCount + uint8(pt)) }

// Type returns the piece kind. It is only meaningful when p is not NoPiece.
func (p Piece) Type() PieceType { return PieceType(uint8(p) % PieceTypeCount) }

// Color returns the owning side. It is only meaningful when p is not NoPiece.
func (p Piece) Color() Color { return Color(uint8(p) / PieceTypeCount) }

// String returns the FEN letter for the piece, or "." for NoPiece.
func (p Piece) String() string {
	if p >= NoPiece {
		return "."
	}
	return string(rune(pieceChars[p]))
}

// PieceFromChar parses a FEN piece letter, returning NoPiece for anything else.
func PieceFromChar(c byte) Piece {
	if i := strings.IndexByte(pieceChars, c); i >= 0 {
		return Piece(i)
	}
	return NoPiece
}
