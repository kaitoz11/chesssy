package search

import "github.com/kaitoz11/chesssy/engine"

// negamax searches the tree below the current position and returns its score from
// the point of view of the side to move.
//
// It is fail-soft: when the true score falls outside the window alpha..beta the
// best score found is still returned rather than the window bound, which makes the
// value useful to the caller and to the transposition table.
//
// The search is principal variation search: the first move is searched with the
// full window, and every later move with a zero-width window that only asks "is
// this better than the best so far?". That question is much cheaper to answer, and
// a move that unexpectedly answers yes is re-searched properly.
//
// pvNode marks the nodes on the principal variation, where the search is careful:
// no transposition cutoffs and no eval-based pruning, because those trade accuracy
// for speed and the principal variation is where accuracy matters.
func (w *worker) negamax(alpha, beta, depth, ply int, pvNode bool) int {
	p := w.pos
	w.pvLength[ply] = 0

	inCheck := p.InCheck()
	// A position where the king is attacked must never be evaluated statically, so
	// spend another ply resolving the check first.
	if inCheck && depth < MaxPly-2 {
		depth++
	}
	if depth <= 0 {
		return w.quiescence(alpha, beta, ply)
	}

	w.nodes++
	if w.nodes%nodesBetweenLimitChecks == 0 {
		w.checkLimits()
	}
	if w.stopped() {
		return 0
	}

	rootNode := ply == 0
	if !rootNode {
		if p.IsDraw() {
			return drawScore
		}
		if ply >= MaxPly-1 {
			return w.evaluate(p)
		}
		if done, score := mateDistancePruning(&alpha, &beta, ply); done {
			return score
		}
	}
	w.selDepth = max(w.selDepth, ply)

	entry, cutoff, slot := w.lookup(alpha, beta, depth, ply, pvNode)
	if cutoff {
		return entry.Score
	}

	// With no table move to try first, ordering at this node is guesswork, so spend a
	// ply less on it. The move this search does find is left behind in the table, and
	// the next visit is ordered by it.
	if depth >= iirMinDepth && entry.Move == engine.NoMove && !inCheck {
		depth--
	}

	eval, evalExact := w.staticEval(entry, inCheck, alpha, beta)
	w.evalStack[ply] = eval
	improving := w.improving(ply, eval, inCheck)

	// Whole-node pruning: these skip the node entirely, so they are only allowed
	// away from the principal variation, out of check, and with no mate in play.
	if !pvNode && !inCheck && !isMate(beta) {
		if score, ok := w.reverseFutilityPruning(eval, beta, depth, improving); ok {
			return score
		}
		if score, ok := w.nullMovePruning(eval, beta, depth, ply); ok {
			return score
		}
	}

	list := &w.lists[ply]
	p.GenerateMoves(list)
	if list.Len() == 0 {
		// No legal move: mate when in check, stalemate otherwise. Scoring the mate
		// by ply is what makes the search prefer the shortest one.
		if inCheck {
			return -mateScore + ply
		}
		return drawScore
	}

	picker := &w.pickers[ply]
	picker.init(w, list, entry.Move, ply)

	// Squares from which a move would check the enemy king. Checking moves are
	// never pruned or reduced, because forced mates run through quiet checks.
	checkSquares := p.CheckSquares()

	var (
		triedQuiets   [maxTrackedQuiets]engine.Move
		quietCount    int
		moveCount     int
		bestScore     = -infinity
		bestMove      = engine.NoMove
		originalAlpha = alpha
	)

	for m := picker.next(); m != engine.NoMove; m = picker.next() {
		moveCount++
		isQuiet := p.IsQuiet(m)
		givesCheck := m.Kind() != engine.CastlingMove &&
			checkSquares[p.PieceAt(m.From()).Type()].Has(m.To())

		// Per-move pruning waits until one move has a real score, which guarantees
		// that every node searches at least one move.
		if !rootNode && bestScore > -mateInMaxPly && !inCheck && !givesCheck {
			if w.pruneMove(m, alpha, eval, depth, moveCount, isQuiet, improving) {
				continue
			}
		}

		reduction := w.reduction(m, depth, moveCount, isQuiet, givesCheck, pvNode, improving)

		w.moveStack[ply] = m
		p.MakeMove(m)
		score := w.searchChild(alpha, beta, depth, ply, moveCount, pvNode, reduction)
		p.UnmakeMove()

		if w.stopped() {
			return 0
		}

		if score > bestScore {
			bestScore, bestMove = score, m
			if score > alpha {
				alpha = score
				if rootNode {
					w.rootBest = m
				}
				w.recordPV(m, ply)

				if score >= beta {
					// Beta cutoff: the opponent would avoid this line, so there is
					// nothing to gain from searching the rest.
					if isQuiet {
						w.rewardCutoff(m, triedQuiets[:quietCount], depth, ply)
					}
					break
				}
			}
		}

		if isQuiet && quietCount < len(triedQuiets) {
			triedQuiets[quietCount] = m
			quietCount++
		}
	}

	w.tt.Store(slot, p.Key(), Entry{
		Move:       bestMove,
		Score:      scoreToTT(bestScore, ply),
		StaticEval: storableEval(eval, evalExact),
		Depth:      depth,
		Bound:      boundFor(bestScore, originalAlpha, beta),
	})
	return bestScore
}

// maxTrackedQuiets bounds how many quiet moves are remembered for the history
// penalty. Nodes rarely try more before cutting off, and the tail carries little
// information.
const maxTrackedQuiets = 32

// searchChild searches the position after a move, applying principal variation
// search: a full window for the first move, and a zero-width scout window for the
// rest, re-searched when it comes back above alpha.
func (w *worker) searchChild(alpha, beta, depth, ply, moveCount int, pvNode bool, reduction int) int {
	newDepth := depth - 1
	if moveCount == 1 {
		return -w.negamax(-beta, -alpha, newDepth, ply+1, pvNode)
	}

	score := -w.negamax(-alpha-1, -alpha, newDepth-reduction, ply+1, false)
	// A reduced search that beat alpha may have been reduced too far; verify it at
	// full depth before trusting it.
	if score > alpha && reduction > 0 {
		score = -w.negamax(-alpha-1, -alpha, newDepth, ply+1, false)
	}
	// Still above alpha and inside the window: this move may be the new best, so
	// search it properly to get an exact score and a principal variation.
	if score > alpha && score < beta {
		score = -w.negamax(-beta, -alpha, newDepth, ply+1, true)
	}
	return score
}

// lookup probes the transposition table and reports whether its score can stand in for
// a search of this node. The entry is returned either way, since its move is worth
// trying first even when its score is not usable, and so is the slot, which is where
// this node's own result will go.
func (w *worker) lookup(alpha, beta, depth, ply int, pvNode bool) (Entry, bool, Slot) {
	entry, found, slot := w.tt.Probe(w.pos.Key())
	if !found {
		return Entry{Move: engine.NoMove}, false, slot
	}
	entry.Score = scoreFromTT(entry.Score, ply)

	// Only a search at least as deep is trustworthy, and never on the principal
	// variation, where an exact line is wanted rather than a bound.
	usable := !pvNode && entry.Depth >= depth && entry.allowsCutoff(alpha, beta)
	return entry, usable, slot
}

// staticEval returns the static evaluation of the current position, reusing the one
// the transposition table recorded when it has it. In check there is no meaningful
// static score, so it returns noEval.
func (w *worker) staticEval(entry Entry, inCheck bool, alpha, beta int) (score int, exact bool) {
	switch {
	case inCheck:
		return noEval, false
	case entry.Bound != BoundNone && entry.StaticEval != noEval:
		return entry.StaticEval, true
	default:
		return evaluateLazy(w.pos, w.pawns, alpha, beta)
	}
}

// storableEval returns the value to record in the transposition table: a shortcut score
// is not stored, so that a later visit does not mistake it for the full evaluation.
func storableEval(eval int, exact bool) int {
	if !exact {
		return noEval
	}
	return eval
}

// improving reports whether our static score has risen since our last turn. When it
// has, the position is trending our way and the search can afford to prune more.
func (w *worker) improving(ply, eval int, inCheck bool) bool {
	return !inCheck && ply >= 2 && w.evalStack[ply-2] != noEval && eval > w.evalStack[ply-2]
}

// boundFor classifies a node's result for the transposition table.
func boundFor(bestScore, originalAlpha, beta int) Bound {
	switch {
	case bestScore <= originalAlpha:
		return BoundUpper // no move reached alpha; the true score is at most this
	case bestScore >= beta:
		return BoundLower // a move beat beta; the true score is at least this
	default:
		return BoundExact
	}
}
