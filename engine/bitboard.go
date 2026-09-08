package engine

import (
	"math/bits"
	"strings"
)

// Bitboard is a set of squares, one bit per square, with bit 0 = A1.
type Bitboard uint64

const (
	EmptyBB Bitboard = 0
	FullBB  Bitboard = ^Bitboard(0)

	FileABB Bitboard = 0x0101010101010101
	FileBBB Bitboard = FileABB << 1
	FileCBB Bitboard = FileABB << 2
	FileDBB Bitboard = FileABB << 3
	FileEBB Bitboard = FileABB << 4
	FileFBB Bitboard = FileABB << 5
	FileGBB Bitboard = FileABB << 6
	FileHBB Bitboard = FileABB << 7

	Rank1BB Bitboard = 0xFF
	Rank2BB Bitboard = Rank1BB << (8 * 1)
	Rank3BB Bitboard = Rank1BB << (8 * 2)
	Rank4BB Bitboard = Rank1BB << (8 * 3)
	Rank5BB Bitboard = Rank1BB << (8 * 4)
	Rank6BB Bitboard = Rank1BB << (8 * 5)
	Rank7BB Bitboard = Rank1BB << (8 * 6)
	Rank8BB Bitboard = Rank1BB << (8 * 7)

	DarkSquaresBB  Bitboard = 0xAA55AA55AA55AA55
	LightSquaresBB Bitboard = ^DarkSquaresBB
)

// SquareBB returns a bitboard holding only s.
func SquareBB(s Square) Bitboard { return 1 << s }

// FileBB returns every square on file f.
func FileBB(f File) Bitboard { return FileABB << f }

// RankBB returns every square on rank r.
func RankBB(r Rank) Bitboard { return Rank1BB << (8 * r) }

// RelativeRankBB returns rank r seen from c's point of view, so
// RelativeRankBB(Black, Rank2BB) is rank 7.
func RelativeRankBB(c Color, bb Bitboard) Bitboard {
	if c == White {
		return bb
	}
	return bb.Mirror()
}

// Has reports whether s is a member of b.
func (b Bitboard) Has(s Square) bool { return b&SquareBB(s) != 0 }

// With returns b plus s.
func (b Bitboard) With(s Square) Bitboard { return b | SquareBB(s) }

// Without returns b minus s.
func (b Bitboard) Without(s Square) Bitboard { return b & ^SquareBB(s) }

// Count returns the number of squares in b.
func (b Bitboard) Count() int { return bits.OnesCount64(uint64(b)) }

// IsEmpty reports whether b holds no squares.
func (b Bitboard) IsEmpty() bool { return b == 0 }

// MoreThanOne reports whether b holds at least two squares.
func (b Bitboard) MoreThanOne() bool { return b&(b-1) != 0 }

// LSB returns the lowest square in b. b must not be empty.
func (b Bitboard) LSB() Square { return Square(bits.TrailingZeros64(uint64(b))) }

// PopLSB removes and returns the lowest square in b. b must not be empty.
func (b *Bitboard) PopLSB() Square {
	s := b.LSB()
	*b &= *b - 1
	return s
}

// Mirror flips the bitboard vertically, swapping rank 1 with rank 8.
func (b Bitboard) Mirror() Bitboard { return Bitboard(bits.ReverseBytes64(uint64(b))) }

// Shift moves every square in b one step in direction d, dropping bits that
// would wrap around the board edges.
func (b Bitboard) Shift(d Direction) Bitboard {
	switch d {
	case North:
		return b << 8
	case South:
		return b >> 8
	case East:
		return (b & ^FileHBB) << 1
	case West:
		return (b & ^FileABB) >> 1
	case NorthEast:
		return (b & ^FileHBB) << 9
	case NorthWest:
		return (b & ^FileABB) << 7
	case SouthEast:
		return (b & ^FileHBB) >> 7
	case SouthWest:
		return (b & ^FileABB) >> 9
	default:
		return EmptyBB
	}
}

// String renders the bitboard as an 8x8 grid with rank 8 on top.
func (b Bitboard) String() string {
	var sb strings.Builder
	for r := int(Rank8); r >= int(Rank1); r-- {
		sb.WriteByte('1' + byte(r))
		sb.WriteByte(':')
		for f := FileA; f <= FileH; f++ {
			sb.WriteByte(' ')
			if b.Has(NewSquare(f, Rank(r))) {
				sb.WriteByte('x')
			} else {
				sb.WriteByte('.')
			}
		}
		sb.WriteByte('\n')
	}
	sb.WriteString("   a b c d e f g h\n")
	return sb.String()
}
