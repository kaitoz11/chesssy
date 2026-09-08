// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package search

import (
	"time"

	"github.com/kaitoz11/chesssy/engine"
)

// Iterative deepening searches to depth 1, then 2, and so on. That sounds wasteful
// and is the opposite: each iteration leaves behind a transposition table full of
// best moves and a warm set of ordering heuristics, so the next one finds its
// cutoffs almost immediately. It also means there is always a move ready when the
// clock runs out.

// Info reports what the search has found. It is delivered once per completed
// iteration through [Searcher.OnInfo], and once more as the return value of
// [Searcher.Search].
type Info struct {
	// Depth is the iteration this result comes from; SelDepth is the deepest ply
	// actually visited, which is larger because of extensions and quiescence.
	Depth    int
	SelDepth int

	// Score is in centipawns from the point of view of the side to move. When the
	// search proved a mate, Mate holds its distance in moves, positive when the
	// side to move mates and negative when it is mated.
	Score int
	Mate  int

	Nodes   uint64
	Elapsed time.Duration

	// PV is the principal variation: the line the search expects to be played.
	PV []engine.Move
}

// BestMove returns the move to play, or engine.NoMove when the search found none.
func (i Info) BestMove() engine.Move {
	if len(i.PV) == 0 {
		return engine.NoMove
	}
	return i.PV[0]
}

// NodesPerSecond returns the search speed, the usual measure of engine throughput.
func (i Info) NodesPerSecond() uint64 {
	if i.Elapsed <= 0 {
		return 0
	}
	return uint64(float64(i.Nodes) / i.Elapsed.Seconds())
}

// iterate deepens until a limit stops it, reporting each completed iteration.
func (w *worker) iterate() Info {
	maxDepth := w.limits.Depth
	if maxDepth <= 0 || maxDepth > MaxPly-2 {
		maxDepth = MaxPly - 2
	}

	var best Info
	score := 0
	for depth := w.startDepth(); depth <= maxDepth; depth++ {
		if w.skipDepth(depth) {
			continue
		}
		score = w.searchRoot(depth, score)

		// An aborted iteration is incomplete, so keep the previous result. The
		// first iteration is the exception: a partial answer beats none.
		if w.stopped() && best.Depth > 0 {
			break
		}

		best = w.collectInfo(depth, score)
		if w.onInfo != nil {
			w.onInfo(best)
		}

		if w.stopped() || w.outOfTimeForNextIteration() {
			break
		}
	}
	return best
}

// collectInfo snapshots the state of the search after an iteration.
func (w *worker) collectInfo(depth, score int) Info {
	// Publish the exact count: between limit checks the shared counter lags, and an
	// iteration boundary is where the number is reported.
	w.published.Store(w.nodes)

	info := Info{
		Depth:    depth,
		SelDepth: w.selDepth,
		Score:    score,
		Mate:     mateDistance(score),
		Nodes:    w.nodeCount(),
		Elapsed:  time.Since(w.start),
		PV:       w.principalVariation(),
	}
	// An iteration cut short can leave the table empty while still having found a
	// move worth playing.
	if len(info.PV) == 0 && w.rootBest != engine.NoMove {
		info.PV = []engine.Move{w.rootBest}
	}
	return info
}

// searchRoot searches one iteration, narrowing the window around the previous
// score. A narrow window prunes far more, at the risk of the true score falling
// outside it; when that happens the window is widened and the iteration repeats.
func (w *worker) searchRoot(depth, previous int) int {
	alpha, beta, delta := -infinity, infinity, aspirationDelta
	if depth >= aspirationMinDepth && !isMate(previous) {
		alpha, beta = previous-delta, previous+delta
	}

	for {
		score := w.negamax(alpha, beta, depth, 0, true)
		if w.stopped() {
			return score
		}
		switch {
		case score <= alpha: // fail low: the position is worse than assumed
			beta = (alpha + beta) / 2
			alpha = max(score-delta, -infinity)
		case score >= beta: // fail high: better than assumed
			beta = min(score+delta, infinity)
		default:
			return score
		}
		delta += delta/2 + 5
	}
}

// recordPV records m as the best move at ply, appending the line the child found.
func (w *worker) recordPV(m engine.Move, ply int) {
	w.pv[ply][0] = m
	copy(w.pv[ply][1:], w.pv[ply+1][:w.pvLength[ply+1]])
	w.pvLength[ply] = w.pvLength[ply+1] + 1
}

// principalVariation returns a copy of the line from the root.
func (w *worker) principalVariation() []engine.Move {
	pv := make([]engine.Move, w.pvLength[0])
	copy(pv, w.pv[0][:w.pvLength[0]])
	return pv
}
