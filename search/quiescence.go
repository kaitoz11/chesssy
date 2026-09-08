package search

import "github.com/kaitoz11/chesssy/engine"

// quiescence is the search that runs at the leaves. Stopping a search in the middle
// of a capture sequence would evaluate a position where a queen is hanging as
// though the material were settled, so instead of evaluating immediately, the
// leaves keep searching forcing moves until the position is quiet.
//
// The side to move is never forced to capture, so the static evaluation acts as a
// lower bound, the "stand pat" score: if simply standing still is already good
// enough to cut off, nothing needs searching.
func (w *worker) quiescence(alpha, beta, ply int) int {
	p := w.pos

	w.nodes++
	if w.nodes%nodesBetweenLimitChecks == 0 {
		w.checkLimits()
	}
	if w.stopped() {
		return 0
	}
	if p.IsDraw() {
		return drawScore
	}
	if ply >= MaxPly-1 {
		return w.evaluate(p)
	}
	w.selDepth = max(w.selDepth, ply)

	inCheck := p.InCheck()
	entry, cutoff, slot := w.lookupQuiescence(alpha, beta, ply)
	if cutoff {
		return entry.Score
	}

	eval, evalExact := w.staticEval(entry, inCheck, alpha, beta)
	bestScore := -infinity
	if !inCheck {
		// Stand pat. In check this is skipped: standing still is not an option, so
		// the score has to come from a move.
		bestScore = eval
		if bestScore >= beta {
			return bestScore
		}
		alpha = max(alpha, bestScore)
	}

	list := &w.lists[ply]
	p.GenerateCaptures(list)
	if list.Len() == 0 {
		// In check the generator produced every evasion, so an empty list is mate.
		if inCheck {
			return -mateScore + ply
		}
		return bestScore
	}

	picker := &w.pickers[ply]
	picker.init(w, list, entry.Move, ply)

	bestMove := engine.NoMove
	originalAlpha := alpha

	for m := picker.next(); m != engine.NoMove; m = picker.next() {
		// Out of check, a capture only interests us if it can raise alpha. In check
		// every evasion has to be searched, however bad it looks.
		if !inCheck && w.pruneCapture(m, eval, alpha) {
			continue
		}

		w.moveStack[ply] = m
		p.MakeMove(m)
		score := -w.quiescence(-beta, -alpha, ply+1)
		p.UnmakeMove()

		if w.stopped() {
			return 0
		}
		if score > bestScore {
			bestScore, bestMove = score, m
			if score > alpha {
				alpha = score
				if score >= beta {
					break
				}
			}
		}
	}

	w.tt.Store(slot, p.Key(), Entry{
		Move:       bestMove,
		Score:      scoreToTT(bestScore, ply),
		StaticEval: storableEval(eval, evalExact),
		Depth:      0, // quiescence results are only valid at the leaves
		Bound:      boundFor(bestScore, originalAlpha, beta),
	})
	return bestScore
}

// lookupQuiescence probes the transposition table for a leaf. Any depth will do
// here, since quiescence itself searches no deeper.
func (w *worker) lookupQuiescence(alpha, beta, ply int) (Entry, bool, Slot) {
	entry, found, slot := w.tt.Probe(w.pos.Key())
	if !found {
		return Entry{Move: engine.NoMove}, false, slot
	}
	entry.Score = scoreFromTT(entry.Score, ply)
	return entry, entry.allowsCutoff(alpha, beta), slot
}

// pruneCapture reports whether a capture can be skipped in a quiet position.
func (w *worker) pruneCapture(m engine.Move, eval, alpha int) bool {
	// A capture that loses material by static exchange evaluation cannot improve a
	// position that is already settled.
	if !w.pos.SeeGE(m, 0) {
		return true
	}
	// Delta pruning: even winning the captured piece outright would leave the score
	// short of alpha, so the line cannot matter.
	return eval != noEval && eval+captureGain(w.pos, m)+quiescenceDeltaMargin < alpha
}

// captureGain returns the material a capture wins, promotions included.
func captureGain(p *engine.Position, m engine.Move) int {
	gain := 0
	if captured := p.PieceAt(m.To()); captured != engine.NoPiece {
		gain = engine.SeeValue(captured.Type())
	} else if m.Kind() == engine.EnPassantMove {
		gain = engine.SeeValue(engine.Pawn)
	}
	if m.IsPromotion() {
		gain += engine.SeeValue(m.Promotion()) - engine.SeeValue(engine.Pawn)
	}
	return gain
}
