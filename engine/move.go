// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package engine

// Move packs a move into 16 bits:
//
//	bits  0-5   origin square
//	bits  6-11  destination square
//	bits 12-13  promotion piece, counted from Knight
//	bits 14-15  move kind
//
// Castling is encoded as the king's origin and destination, which matches UCI
// notation (e1g1 for white kingside).
type Move uint16

// MoveKind distinguishes the four move encodings. Everything the board needs to
// know beyond "from" and "to" follows from the kind.
type MoveKind uint16

const (
	// NormalMove is a quiet move or a plain capture.
	NormalMove MoveKind = iota
	// PromotionMove is a pawn reaching the last rank, with or without a capture.
	PromotionMove
	// EnPassantMove is a pawn capturing onto an empty square.
	EnPassantMove
	// CastlingMove moves both the king and the rook.
	CastlingMove
)

// NoMove is the zero Move. It is unambiguous as a sentinel because a1a1 can
// never be legal.
const NoMove Move = 0

// Bit layout of a Move.
const (
	moveFromShift  = 0
	moveToShift    = 6
	movePromoShift = 12
	moveKindShift  = 14
	moveSquareMask = 0x3F
	movePromoMask  = 0x3
)

// NewMove builds a quiet move or a plain capture.
func NewMove(from, to Square) Move {
	return Move(uint16(from)<<moveFromShift | uint16(to)<<moveToShift)
}

// NewSpecialMove builds an en passant capture or a castling move.
func NewSpecialMove(from, to Square, kind MoveKind) Move {
	return NewMove(from, to) | Move(uint16(kind)<<moveKindShift)
}

// NewPromotion builds a promotion to promo, with or without a capture.
func NewPromotion(from, to Square, promo PieceType) Move {
	return NewMove(from, to) |
		Move(uint16(promo-Knight)<<movePromoShift) |
		Move(uint16(PromotionMove)<<moveKindShift)
}

// From returns the origin square.
func (m Move) From() Square { return Square(m>>moveFromShift) & moveSquareMask }

// To returns the destination square.
func (m Move) To() Square { return Square(m>>moveToShift) & moveSquareMask }

// Kind returns how the move is encoded.
func (m Move) Kind() MoveKind { return MoveKind(m >> moveKindShift) }

// Promotion returns the piece the pawn becomes. It is only meaningful when Kind
// is PromotionMove.
func (m Move) Promotion() PieceType {
	return PieceType(m>>movePromoShift&movePromoMask) + Knight
}

// IsPromotion reports whether the move promotes a pawn.
func (m Move) IsPromotion() bool { return m.Kind() == PromotionMove }

// String renders the move in UCI long algebraic notation, such as "e2e4" or
// "e7e8q". NoMove renders as "0000", the value UCI uses for "no move".
func (m Move) String() string {
	if m == NoMove {
		return "0000"
	}
	s := m.From().String() + m.To().String()
	if m.IsPromotion() {
		s += m.Promotion().String()
	}
	return s
}
