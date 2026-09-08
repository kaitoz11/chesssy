// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package engine

// random is a xorshift64* generator. The engine needs reproducible pseudo-random
// numbers for the magic search and the Zobrist keys, and it needs them before
// math/rand is a good fit: seeded once, identical on every run, no locking.
type random struct{ state uint64 }

// newRandom returns a generator seeded with seed. A zero seed is replaced, since
// xorshift cannot escape zero.
func newRandom(seed uint64) *random {
	if seed == 0 {
		seed = 0x9E3779B97F4A7C15
	}
	return &random{state: seed}
}

// uint64 returns the next value in the sequence.
func (r *random) uint64() uint64 {
	r.state ^= r.state >> 12
	r.state ^= r.state << 25
	r.state ^= r.state >> 27
	return r.state * 2685821657736338717
}

// sparseUint64 returns a value with roughly an eighth of its bits set. Sparse
// values make far better magic multipliers than uniformly random ones.
func (r *random) sparseUint64() uint64 {
	// Three independent draws combined: a bit survives only if it is set in all
	// three, so each has about a one in eight chance.
	a, b, c := r.uint64(), r.uint64(), r.uint64()
	return a & b & c
}
