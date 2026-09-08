package engine

import (
	"testing"
)

// TestMagicAttacksMatchReference checks every magic lookup against the slow ray
// walker for a large number of random occupancies. If a magic multiplier had a
// destructive collision, this is where it would show up.
func TestMagicAttacksMatchReference(t *testing.T) {
	rng := newRandom(0xC0FFEE)

	for s := A1; s < SquareCount; s++ {
		for i := 0; i < 2000; i++ {
			// Two draws ANDed: a sparse occupancy leaves more squares reachable,
			// which exercises longer rays than a uniformly random board would.
			a, b := rng.uint64(), rng.uint64()
			occupied := Bitboard(a & b)

			if got, want := RookAttacks(s, occupied), slowSlidingAttacks(rookDirections, s, occupied); got != want {
				t.Fatalf("RookAttacks(%v, %#x):\ngot:\n%v\nwant:\n%v", s, occupied, got, want)
			}
			if got, want := BishopAttacks(s, occupied), slowSlidingAttacks(bishopDirections, s, occupied); got != want {
				t.Fatalf("BishopAttacks(%v, %#x):\ngot:\n%v\nwant:\n%v", s, occupied, got, want)
			}
		}
	}
}

func TestMagicAttacksIgnoreIrrelevantOccupancy(t *testing.T) {
	// Occupancy outside the relevant mask must not change the result: edge
	// squares carry no blocking information.
	for s := A1; s < SquareCount; s++ {
		for _, m := range []*magic{&rookMagics[s], &bishopMagics[s]} {
			outside := ^m.mask
			if m.attacks[m.index(EmptyBB)] != m.attacks[m.index(outside)] {
				t.Fatalf("square %v: attacks changed with occupancy outside the mask", s)
			}
		}
	}
}

func BenchmarkRookAttacks(b *testing.B) {
	occupied := Rank2BB | Rank7BB
	var sink Bitboard
	for i := 0; i < b.N; i++ {
		sink |= RookAttacks(Square(i&63), occupied)
	}
	if sink == 0 {
		b.Fatal("unreachable")
	}
}
