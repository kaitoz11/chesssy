// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package search

import "github.com/kaitoz11/chesssy/engine"

// Pawn structure is the slowest part of the evaluation and the slowest to change:
// most moves leave the pawns exactly where they were. So it is evaluated on its own
// and memoised, keyed by the pawn placement of both sides.

// Pawn term weights, {middlegame, endgame}.
var (
	// A pawn behind a friendly pawn on the same file cannot support an attack and
	// is hard to defend.
	doubledPawn = [2]int{-11, -26}
	// A pawn with no friendly pawn on an adjacent file can never be defended by
	// one.
	isolatedPawn = [2]int{-14, -12}
	// A pawn that cannot be defended by a pawn and cannot safely advance.
	backwardPawn = [2]int{-8, -10}
	// A pawn with no enemy pawn ahead of it on its own or an adjacent file. The
	// bonus grows sharply as it nears promotion, and counts for far more in the
	// endgame, where there is less to stop it.
	passedPawnByRank = [2][engine.RankCount]int{
		{0, 5, 10, 22, 42, 75, 120, 0},
		{0, 12, 22, 40, 72, 120, 175, 0},
	}
)

// Precomputed masks. Each is a pure function of the square, so they are built once
// as package variables; the compiler orders them by dependency.
var (
	// adjacentFiles[f] holds the files either side of f.
	adjacentFiles = computeAdjacentFiles()
	// frontSpan[c][s] holds the squares ahead of s on its own file, from c's point
	// of view.
	frontSpan = computeFrontSpans()
	// passedSpan[c][s] holds the squares ahead of s on its own and adjacent files.
	// A pawn with no enemy pawn in that region is passed.
	passedSpan = computePassedSpans()
)

func computeAdjacentFiles() (files [engine.FileCount]engine.Bitboard) {
	for f := engine.FileA; f <= engine.FileH; f++ {
		if f > engine.FileA {
			files[f] |= engine.FileBB(f - 1)
		}
		if f < engine.FileH {
			files[f] |= engine.FileBB(f + 1)
		}
	}
	return files
}

func computeFrontSpans() (spans [engine.ColorCount][engine.SquareCount]engine.Bitboard) {
	for s := engine.A1; s < engine.SquareCount; s++ {
		file := engine.FileBB(s.File())
		spans[engine.White][s] = ranksAhead(engine.White, s.Rank()) & file
		spans[engine.Black][s] = ranksAhead(engine.Black, s.Rank()) & file
	}
	return spans
}

func computePassedSpans() (spans [engine.ColorCount][engine.SquareCount]engine.Bitboard) {
	for s := engine.A1; s < engine.SquareCount; s++ {
		region := engine.FileBB(s.File()) | adjacentFiles[s.File()]
		spans[engine.White][s] = ranksAhead(engine.White, s.Rank()) & region
		spans[engine.Black][s] = ranksAhead(engine.Black, s.Rank()) & region
	}
	return spans
}

// ranksAhead returns every rank in front of r from c's point of view.
func ranksAhead(c engine.Color, r engine.Rank) engine.Bitboard {
	var bb engine.Bitboard
	if c == engine.White {
		for ahead := r + 1; ahead <= engine.Rank8; ahead++ {
			bb |= engine.RankBB(ahead)
		}
		return bb
	}
	for behind := engine.Rank1; behind < r; behind++ {
		bb |= engine.RankBB(behind)
	}
	return bb
}

// pawnCache memoises the pawn-only score. It belongs to a Searcher rather than being
// global, so that concurrent searchers stay independent.
type pawnCache struct {
	entries [pawnCacheSize]pawnCacheEntry
}

// pawnCacheEntry stores the pawn placement it was computed for, which is both the
// verification and the key: two positions with the same pawns have the same pawn
// score, whatever else differs.
type pawnCacheEntry struct {
	white, black engine.Bitboard
	mg, eg       int32
}

const (
	pawnCacheBits = 13
	pawnCacheSize = 1 << pawnCacheBits
)

// pawnScore returns the pawn placement and structure score, white minus black,
// consulting cache when one is given.
func pawnScore(pawns [engine.ColorCount]engine.Bitboard, cache *pawnCache) (mg, eg int) {
	if cache == nil {
		return computePawnScore(pawns)
	}

	// A zero entry only matches a board with no pawns, whose score is zero anyway,
	// so no separate validity flag is needed.
	e := &cache.entries[pawnCacheIndex(pawns)]
	if e.white == pawns[engine.White] && e.black == pawns[engine.Black] {
		return int(e.mg), int(e.eg)
	}

	mg, eg = computePawnScore(pawns)
	e.white, e.black = pawns[engine.White], pawns[engine.Black]
	e.mg, e.eg = int32(mg), int32(eg)
	return mg, eg
}

// pawnCacheIndex hashes both pawn sets into a slot.
func pawnCacheIndex(pawns [engine.ColorCount]engine.Bitboard) uint64 {
	const (
		whiteMultiplier = 0x9E3779B97F4A7C15
		blackMultiplier = 0xC2B2AE3D27D4EB4F
	)
	h := uint64(pawns[engine.White])*whiteMultiplier ^ uint64(pawns[engine.Black])*blackMultiplier
	return (h >> 32) & (pawnCacheSize - 1)
}

// computePawnScore sums the placement and structure terms of every pawn, white
// counted positive and black negative.
func computePawnScore(pawns [engine.ColorCount]engine.Bitboard) (mg, eg int) {
	for c := engine.White; c < engine.ColorCount; c++ {
		sign := signFor(c)
		piece := engine.NewPiece(c, engine.Pawn)

		for bb := pawns[c]; bb != 0; {
			s := bb.PopLSB()
			structureMg, structureEg := pawnTerms(c, s, pawns)
			mg += sign * (mgTable[piece][s] + structureMg)
			eg += sign * (egTable[piece][s] + structureEg)
		}
	}
	return mg, eg
}

// pawnTerms scores one pawn's structural strengths and weaknesses.
func pawnTerms(c engine.Color, s engine.Square, pawns [engine.ColorCount]engine.Bitboard) (mg, eg int) {
	own, enemy := pawns[c], pawns[c.Flip()]
	file := s.File()

	if frontSpan[c][s]&own != 0 {
		mg += doubledPawn[0]
		eg += doubledPawn[1]
	}

	switch {
	case own&adjacentFiles[file] == 0:
		mg += isolatedPawn[0]
		eg += isolatedPawn[1]

	case engine.PawnAttacksBB(c.Flip(), engine.SquareBB(s))&own == 0 && frontSpan[c][s]&enemy == 0:
		// No friendly pawn can ever defend this one, and nothing blocks the file
		// for the enemy to attack along.
		mg += backwardPawn[0]
		eg += backwardPawn[1]
	}

	if passedSpan[c][s]&enemy == 0 {
		r := relativeRank(c, s.Rank())
		mg += passedPawnByRank[0][r]
		eg += passedPawnByRank[1][r]
	}
	return mg, eg
}

// relativeRank returns the rank as the given color counts it, so that rank 7 is
// always one step from promotion.
func relativeRank(c engine.Color, r engine.Rank) engine.Rank {
	if c == engine.White {
		return r
	}
	return engine.Rank8 - r
}

// signFor returns +1 for white and -1 for black, so that a single loop can add both
// sides into one score.
func signFor(c engine.Color) int {
	if c == engine.White {
		return 1
	}
	return -1
}
