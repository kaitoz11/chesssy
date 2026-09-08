// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package engine

// Attack tables. Leapers (pawn, knight, king) get one bitboard per square.
// Sliders go through magic.go. The between and line tables answer the geometry
// questions move generation asks about pins and checks.
//
// Every table is built once at package initialisation and never written again, so
// lookups are safe for concurrent use.

var (
	pawnAttackTable   [ColorCount][SquareCount]Bitboard
	knightAttackTable [SquareCount]Bitboard
	kingAttackTable   [SquareCount]Bitboard

	// betweenTable[a][b] holds the squares strictly between a and b when they
	// share a rank, file or diagonal, and nothing otherwise.
	betweenTable [SquareCount][SquareCount]Bitboard
	// lineTable[a][b] holds the whole rank, file or diagonal through a and b.
	lineTable [SquareCount][SquareCount]Bitboard
)

// Magic search seeds. They are arbitrary but fixed, which is what makes the
// generated tables reproducible.
const (
	rookMagicSeed   = 0x2A5F_C1D8_9B47_0E13
	bishopMagicSeed = 0x7C1E_2B9A_4D63_F085
)

func init() {
	initLeaperAttacks()
	initMagics(rookDirections, rookTable[:], &rookMagics, rookMagicSeed)
	initMagics(bishopDirections, bishopTable[:], &bishopMagics, bishopMagicSeed)
	initRays() // needs the slider tables above
}

func initLeaperAttacks() {
	knightSteps := [8][2]int8{{1, 2}, {2, 1}, {2, -1}, {1, -2}, {-1, -2}, {-2, -1}, {-2, 1}, {-1, 2}}
	kingSteps := [8][2]int8{{0, 1}, {1, 1}, {1, 0}, {1, -1}, {0, -1}, {-1, -1}, {-1, 0}, {-1, 1}}

	for s := A1; s < SquareCount; s++ {
		f, r := int8(s.File()), int8(s.Rank())

		for _, step := range knightSteps {
			if to, ok := offsetSquare(f, r, step[0], step[1]); ok {
				knightAttackTable[s] = knightAttackTable[s].With(to)
			}
		}
		for _, step := range kingSteps {
			if to, ok := offsetSquare(f, r, step[0], step[1]); ok {
				kingAttackTable[s] = kingAttackTable[s].With(to)
			}
		}
		// Pawns capture diagonally forward, in opposite directions per color.
		for _, df := range [2]int8{1, -1} {
			if to, ok := offsetSquare(f, r, df, 1); ok {
				pawnAttackTable[White][s] = pawnAttackTable[White][s].With(to)
			}
			if to, ok := offsetSquare(f, r, df, -1); ok {
				pawnAttackTable[Black][s] = pawnAttackTable[Black][s].With(to)
			}
		}
	}
}

// initRays fills the between and line tables by intersecting the rays of a piece
// standing on each square with those of a piece standing on the other.
func initRays() {
	for a := A1; a < SquareCount; a++ {
		for _, directions := range [2][4][2]int8{rookDirections, bishopDirections} {
			raysFromA := slowSlidingAttacks(directions, a, EmptyBB)
			for b := A1; b < SquareCount; b++ {
				if a == b || !raysFromA.Has(b) {
					continue // not aligned along these directions
				}
				lineTable[a][b] = (raysFromA & slowSlidingAttacks(directions, b, EmptyBB)).With(a).With(b)
				// Blocking each ray with the other square leaves exactly the
				// squares between them.
				betweenTable[a][b] = slowSlidingAttacks(directions, a, SquareBB(b)) &
					slowSlidingAttacks(directions, b, SquareBB(a))
			}
		}
	}
}

// offsetSquare returns the square df files and dr ranks away from (f, r), and
// whether it is still on the board.
func offsetSquare(f, r, df, dr int8) (Square, bool) {
	nf, nr := f+df, r+dr
	if nf < 0 || nf > 7 || nr < 0 || nr > 7 {
		return NoSquare, false
	}
	return NewSquare(File(nf), Rank(nr)), true
}

// PawnAttacks returns the squares a pawn of color c on s attacks.
func PawnAttacks(c Color, s Square) Bitboard { return pawnAttackTable[c][s] }

// KnightAttacks returns the squares a knight on s attacks.
func KnightAttacks(s Square) Bitboard { return knightAttackTable[s] }

// KingAttacks returns the squares a king on s attacks.
func KingAttacks(s Square) Bitboard { return kingAttackTable[s] }

// QueenAttacks returns the squares a queen on s attacks given the occupancy.
func QueenAttacks(s Square, occupied Bitboard) Bitboard {
	return RookAttacks(s, occupied) | BishopAttacks(s, occupied)
}

// PieceAttacks returns the squares a piece of type pt on s attacks given the
// occupancy. Pawns are not included, since they capture and move differently;
// use PawnAttacks for those.
func PieceAttacks(pt PieceType, s Square, occupied Bitboard) Bitboard {
	switch pt {
	case Knight:
		return knightAttackTable[s]
	case Bishop:
		return BishopAttacks(s, occupied)
	case Rook:
		return RookAttacks(s, occupied)
	case Queen:
		return QueenAttacks(s, occupied)
	case King:
		return kingAttackTable[s]
	default:
		return EmptyBB
	}
}

// PawnAttacksBB returns every square attacked by the pawns of color c in bb.
func PawnAttacksBB(c Color, bb Bitboard) Bitboard {
	if c == White {
		return bb.Shift(NorthEast) | bb.Shift(NorthWest)
	}
	return bb.Shift(SouthEast) | bb.Shift(SouthWest)
}

// BetweenBB returns the squares strictly between a and b, or nothing when they do
// not share a rank, file or diagonal.
func BetweenBB(a, b Square) Bitboard { return betweenTable[a][b] }

// LineBB returns the whole rank, file or diagonal through a and b, or nothing
// when there is none.
func LineBB(a, b Square) Bitboard { return lineTable[a][b] }

// Aligned reports whether a, b and c lie on one line.
func Aligned(a, b, c Square) bool { return lineTable[a][b].Has(c) }
