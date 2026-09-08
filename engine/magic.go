// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package engine

// Sliding piece attacks use plain magic bitboards. For a given square, only the
// occupancy of the squares the piece could reach matters, and the board edges
// cannot block anything further, so a rook on a1 has just 12 relevant squares.
// Multiplying those bits by a magic constant scatters them into the top of a
// 64-bit word, and the top bits index a dense table of precomputed attack sets.
//
// The magic constants are searched for at package initialisation with a
// fixed-seed generator, so the tables are identical on every run without a
// hard-coded list of magics in the source. TestMagicAttacksMatchReference checks
// every lookup against slowSlidingAttacks.

// Table sizes are the sum over all squares of 1 << (relevant occupancy bits).
const (
	rookTableSize   = 102400
	bishopTableSize = 5248
)

// magic holds the perfect hash for one square.
type magic struct {
	mask    Bitboard   // the relevant occupancy squares, edges excluded
	mul     uint64     // the magic multiplier
	shift   uint8      // 64 minus the number of bits in mask
	attacks []Bitboard // 1 << bits entries, indexed by the hashed occupancy
}

// index hashes an occupancy into a slot in m.attacks.
func (m *magic) index(occupied Bitboard) uint64 {
	return (uint64(occupied&m.mask) * m.mul) >> m.shift
}

var (
	rookMagics   [SquareCount]magic
	bishopMagics [SquareCount]magic
	rookTable    [rookTableSize]Bitboard
	bishopTable  [bishopTableSize]Bitboard
)

// Ray directions as (file, rank) steps.
var (
	rookDirections   = [4][2]int8{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}
	bishopDirections = [4][2]int8{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
)

// RookAttacks returns the squares a rook on s attacks given the occupancy.
func RookAttacks(s Square, occupied Bitboard) Bitboard {
	m := &rookMagics[s]
	return m.attacks[m.index(occupied)]
}

// BishopAttacks returns the squares a bishop on s attacks given the occupancy.
func BishopAttacks(s Square, occupied Bitboard) Bitboard {
	m := &bishopMagics[s]
	return m.attacks[m.index(occupied)]
}

// slowSlidingAttacks walks the rays from s until they leave the board or hit an
// occupied square, which is included as a capture. It builds the magic tables and
// serves as the reference implementation in tests.
func slowSlidingAttacks(directions [4][2]int8, s Square, occupied Bitboard) Bitboard {
	var attacks Bitboard
	f, r := int8(s.File()), int8(s.Rank())
	for _, d := range directions {
		for i := int8(1); ; i++ {
			to, ok := offsetSquare(f, r, d[0]*i, d[1]*i)
			if !ok {
				break
			}
			attacks = attacks.With(to)
			if occupied.Has(to) {
				break
			}
		}
	}
	return attacks
}

// initMagics fills magics and table for one slider kind. It is called once per
// kind at package initialisation.
func initMagics(directions [4][2]int8, table []Bitboard, magics *[SquareCount]magic, seed uint64) {
	// Scratch space for the largest relevant occupancy, a rook in a corner-free
	// square with 12 bits.
	var (
		occupancy [4096]Bitboard
		reference [4096]Bitboard
		epoch     [4096]int
		attempt   int
		offset    int
	)
	rng := newRandom(seed)

	for s := A1; s < SquareCount; s++ {
		m := &magics[s]
		m.mask = slowSlidingAttacks(directions, s, EmptyBB) & ^edgesExcluding(s)
		m.shift = uint8(64 - m.mask.Count())

		// Enumerate every subset of the mask with the carry-rippler trick, and
		// record the attacks each one produces.
		size := 0
		for sub := EmptyBB; ; {
			occupancy[size] = sub
			reference[size] = slowSlidingAttacks(directions, s, sub)
			size++
			sub = (sub - m.mask) & m.mask
			if sub == EmptyBB {
				break
			}
		}
		m.attacks = table[offset : offset+size]
		offset += size

		// Look for a multiplier that maps every subset to a slot holding the
		// right attack set. Collisions are fine when the colliding subsets have
		// identical attacks, which is what makes small tables possible.
		for i := 0; i < size; {
			m.mul = randomMagicCandidate(rng, m.mask)
			attempt++
			for i = 0; i < size; i++ {
				idx := m.index(occupancy[i])
				if epoch[idx] < attempt {
					epoch[idx] = attempt
					m.attacks[idx] = reference[i]
				} else if m.attacks[idx] != reference[i] {
					break // destructive collision, try another multiplier
				}
			}
		}
	}
}

// edgesExcluding returns the board edges that are not on s's own rank or file.
// Those squares can never block a slider before it stops, so leaving them out of
// the occupancy mask keeps the tables small.
func edgesExcluding(s Square) Bitboard {
	return ((Rank1BB | Rank8BB) & ^RankBB(s.Rank())) | ((FileABB | FileHBB) & ^FileBB(s.File()))
}

// randomMagicCandidate draws a sparse multiplier that spreads the mask across the
// top byte of the product. Candidates that fail this cheap test almost never work,
// so filtering them here shortens the search considerably.
func randomMagicCandidate(rng *random, mask Bitboard) uint64 {
	for {
		if mul := rng.sparseUint64(); Bitboard((mul*uint64(mask))>>56).Count() >= 6 {
			return mul
		}
	}
}
