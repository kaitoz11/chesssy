// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package engine

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// Perft counts the leaf nodes of the move tree at the given depth. It is the
// standard correctness test for move generation: the counts for well-known
// positions are published, so any generation bug shows up as a mismatch.
//
// The last ply is counted without playing the moves, which is roughly twice as
// fast as recursing to depth zero.
func (p *Position) Perft(depth int) uint64 {
	if depth <= 0 {
		return 1
	}

	var list MoveList
	p.GenerateMoves(&list)
	if depth == 1 {
		return uint64(list.Len())
	}

	var nodes uint64
	for _, m := range list.Moves() {
		p.MakeMove(m)
		nodes += p.Perft(depth - 1)
		p.UnmakeMove()
	}
	return nodes
}

// PerftDivide reports the node count for each legal move along with their total.
// When a perft total is wrong, comparing a divide against a known-good engine
// points straight at the offending move.
func (p *Position) PerftDivide(depth int) (counts map[string]uint64, total uint64) {
	counts = make(map[string]uint64)

	var list MoveList
	p.GenerateMoves(&list)
	for _, m := range list.Moves() {
		p.MakeMove(m)
		nodes := p.Perft(depth - 1)
		p.UnmakeMove()

		counts[m.String()] = nodes
		total += nodes
	}
	return counts, total
}

// FormatDivide renders a divide result in the conventional one-move-per-line
// layout, sorted by move.
func FormatDivide(counts map[string]uint64, total uint64) string {
	var sb strings.Builder
	for _, m := range slices.Sorted(maps.Keys(counts)) {
		fmt.Fprintf(&sb, "%s: %d\n", m, counts[m])
	}
	fmt.Fprintf(&sb, "\nnodes: %d\n", total)
	return sb.String()
}
