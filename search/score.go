// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package search

// Scores are centipawns from the point of view of the side to move, which is what
// negamax needs: a score is negated when it crosses to the other side.
//
// Mate is scored close to the maximum so that it dominates any material count,
// and the distance to it is subtracted, so that a mate in two outranks a mate in
// five. A score at or beyond mateInMaxPly is therefore a proven mate.
const (
	// infinity is wider than any real score, so it is safe as an initial window
	// bound and as "no score yet".
	infinity = 32001

	// mateScore is being mated at the current ply; -mateScore + ply is the score
	// of a mate ply moves away.
	mateScore = 32000

	// mateInMaxPly is the smallest score that still means a forced mate.
	mateInMaxPly = mateScore - MaxPly

	// drawScore is the score of a drawn position. Contempt would go here.
	drawScore = 0

	// noEval marks "no static evaluation available", which happens in check,
	// where a static score would be meaningless.
	noEval = -32768
)

// isMate reports whether score represents a forced mate for either side.
func isMate(score int) bool { return abs(score) >= mateInMaxPly }

// mateDistance converts a mate score into a signed distance in moves: positive
// when the side to move mates, negative when it is mated, zero when the score is
// not a mate.
func mateDistance(score int) int {
	switch {
	case score >= mateInMaxPly:
		return (mateScore - score + 1) / 2
	case score <= -mateInMaxPly:
		return -(mateScore + score + 1) / 2
	default:
		return 0
	}
}

// abs returns the absolute value of v. Scores fit comfortably in an int, so no
// overflow check is needed.
func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
