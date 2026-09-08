package search

import "github.com/kaitoz11/chesssy/engine"

// Pruning is where a chess search buys its depth. Every function here answers
// "can this be skipped?" with a heuristic that is usually but not always right,
// and each one trades a small risk of missing something for a large saving. They
// share two safeguards: none of them applies on the principal variation, and none
// of them applies to a checking move or to a node in check, which is what keeps
// forced mates visible.
//
// The reductions live here too. A reduction is the softer form of the same bet:
// instead of skipping a move it searches it shallower, and re-searches properly if
// the shallow search suggests the bet was wrong.

// mateDistancePruning narrows the window to the mate scores still achievable at
// this ply. A mate found closer to the root elsewhere makes anything slower here
// irrelevant.
func mateDistancePruning(alpha, beta *int, ply int) (done bool, score int) {
	*alpha = max(*alpha, -mateScore+ply)
	*beta = min(*beta, mateScore-ply-1)
	if *alpha >= *beta {
		return true, *alpha
	}
	return false, 0
}

// reverseFutilityPruning gives up on a node whose static score is so far above beta
// that no plausible reply brings it back. It is the mirror of futility pruning:
// instead of asking whether a move can reach alpha, it asks whether the opponent
// can reach beta.
func (w *worker) reverseFutilityPruning(eval, beta, depth int, improving bool) (int, bool) {
	if depth > reverseFutilityMaxDepth {
		return 0, false
	}
	margin := reverseFutilityMargin * depth
	if improving {
		margin -= reverseFutilityImprovingBonus
	}
	if eval-margin >= beta {
		return eval, true
	}
	return 0, false
}

// nullMovePruning asks what happens if we skip our turn. If the opponent still
// cannot bring the score below beta with a free move, our real move will not help
// them either, and the node can be cut.
//
// It is skipped when only pawns remain, where passing may genuinely be best
// (zugzwang), and after another null move, which would just pass the turn back.
func (w *worker) nullMovePruning(eval, beta, depth, ply int) (int, bool) {
	p := w.pos
	if depth < nullMoveMinDepth || eval < beta ||
		!w.hasNonPawnMaterial(p.SideToMove()) ||
		(ply > 0 && w.moveStack[ply-1] == engine.NoMove) {
		return 0, false
	}

	reduction := nullMoveBaseReduction + depth/nullMoveDepthDivisor +
		min((eval-beta)/nullMoveEvalDivisor, nullMoveMaxEvalBonus)

	w.moveStack[ply] = engine.NoMove
	p.MakeNullMove()
	score := -w.negamax(-beta, -beta+1, depth-reduction, ply+1, false)
	p.UnmakeNullMove()

	if w.stopped() || score < beta {
		return 0, false
	}
	// A mate score from a line that includes a free move is not to be trusted.
	if isMate(score) {
		score = beta
	}
	return score, true
}

// pruneMove reports whether a move can be skipped without searching it. Each test
// is a bet that a shallow indicator predicts a deep result; the caller has already
// established that the node is not in check and that skipping is safe.
func (w *worker) pruneMove(m engine.Move, alpha, eval, depth, moveCount int, isQuiet, improving bool) bool {
	if !isQuiet {
		// A capture that loses material by static exchange evaluation rarely
		// repays a search, and the allowance grows with depth.
		return depth <= captureSeeMaxDepth && !w.pos.SeeGE(m, -captureSeeMargin*depth*depth)
	}
	// Late move pruning: with good ordering, the quiet moves at the tail of the
	// list are unlikely to be better than what has already been searched.
	if moveCount > lateMoveThreshold(depth, improving) {
		return true
	}
	// Futility pruning: the static score is so far below alpha that a quiet move
	// cannot plausibly close the gap.
	return depth <= futilityMaxDepth && eval != noEval &&
		eval+futilityBase+futilityPerDepth*depth <= alpha
}

// reduction returns how many plies to shave off a scout search. Late quiet moves
// are searched shallower on the assumption that ordering put the good moves first;
// the assumption is verified by a re-search whenever a reduced move looks good.
func (w *worker) reduction(m engine.Move, depth, moveCount int, isQuiet, givesCheck, pvNode, improving bool) int {
	if depth < lmrMinDepth || moveCount <= lmrMinMoveCount || !isQuiet || givesCheck {
		return 0
	}

	r := baseReduction(depth, moveCount)
	if !pvNode {
		r++ // away from the principal variation, accuracy matters less
	}
	if improving {
		r-- // the position is trending our way, so look a little harder
	}
	if w.historyScore(w.pos.PieceAt(m.From()), m) > historyReductionThreshold {
		r-- // this move has been good before
	}
	// Never reduce into quiescence: that would skip the rest of the line entirely.
	return min(max(r, 0), depth-2)
}
