package search

import "github.com/kaitoz11/chesssy/engine"

// Three heuristics tell the move picker which quiet moves are worth trying first.
// All are learned during the search and all are approximate, which is fine: a
// wrong guess costs a little time, a right one saves a subtree.
//
//   - Killers: the quiet moves that caused a cutoff at this ply elsewhere in the
//     tree. Positions at the same ply tend to be similar enough that the same move
//     works again.
//   - Counter moves: the reply that refuted this exact move last time it was
//     played, whatever the ply.
//   - History: a table of how often a move caused a cutoff, indexed by the piece that
//     moved and the square it moved to. Indexing by the piece rather than by the square
//     it came from generalises better, since what makes a quiet move good is usually
//     where it lands, and the table is small enough to stay in cache.

const (
	// killerSlots is how many killer moves are remembered per ply.
	killerSlots = 2

	// historyMax bounds the history values. The gravity term in addHistory scales
	// with it, so values approach the bound without reaching it.
	historyMax = 1 << 14

	// historyBonusMax caps the credit a single cutoff can award, so that one deep
	// node cannot dominate the table.
	historyBonusMax = 1 << 13

	// historyBonusScale multiplies depth squared into the bonus: a cutoff found
	// deeper in the tree is better evidence.
	historyBonusScale = 4

	// historyReductionThreshold is the history score above which a quiet move is
	// reduced by one ply less.
	historyReductionThreshold = historyMax / 4
)

// historyScore returns what a quiet move is worth in general.
func (w *worker) historyScore(piece engine.Piece, m engine.Move) int32 {
	return w.history[piece][m.To()]
}

// quietScore is the ordering score of a quiet move: what it is worth in general, plus
// what it is worth as a reply to the move just played.
func (w *worker) quietScore(ply int, piece engine.Piece, m engine.Move) int32 {
	score := w.history[piece][m.To()]
	if prev := w.previousMove(ply); prev != engine.NoMove {
		score += w.continuation[prev.To()][piece][m.To()]
	}
	return score
}

// counterMove returns the move that refuted the move played into this node, or
// engine.NoMove when there is none.
func (w *worker) counterMove(ply int) engine.Move {
	prev := w.previousMove(ply)
	if prev == engine.NoMove {
		return engine.NoMove
	}
	return w.counters[w.pos.PieceAt(prev.To())][prev.To()]
}

// rewardCutoff credits the quiet move that caused a beta cutoff and penalises the
// quiet moves that were tried before it and failed.
func (w *worker) rewardCutoff(best engine.Move, tried []engine.Move, depth, ply int) {
	bonus := int32(min(depth*depth*historyBonusScale, historyBonusMax))

	prev := w.previousMove(ply)
	w.credit(prev, w.pos.PieceAt(best.From()), best, bonus)
	for _, m := range tried {
		if m != best {
			w.credit(prev, w.pos.PieceAt(m.From()), m, -bonus)
		}
	}

	if w.killers[ply][0] != best {
		w.killers[ply][1] = w.killers[ply][0]
		w.killers[ply][0] = best
	}
	if prev := w.previousMove(ply); prev != engine.NoMove {
		w.counters[w.pos.PieceAt(prev.To())][prev.To()] = best
	}
}

// credit applies a bonus to both history tables for one move.
func (w *worker) credit(prev engine.Move, piece engine.Piece, m engine.Move, bonus int32) {
	w.addHistory(piece, m, bonus)
	if prev != engine.NoMove {
		applyBonus(&w.continuation[prev.To()][piece][m.To()], bonus)
	}
}

// addHistory applies a bonus to what a move is worth in general.
func (w *worker) addHistory(piece engine.Piece, m engine.Move, bonus int32) {
	applyBonus(&w.history[piece][m.To()], bonus)
}

// applyBonus adds a bonus with a gravity term: the closer a value is to the bound, the
// less a further bonus moves it, and the more a penalty does. That keeps the tables
// bounded and lets stale entries decay as new evidence arrives.
func applyBonus(h *int32, bonus int32) {
	*h += bonus - *h*abs32(bonus)/historyMax
}

func (w *worker) decayHistory() {
	for piece := range w.history {
		for to := range w.history[piece] {
			w.history[piece][to] /= 2
		}
	}
	for prev := range w.continuation {
		for piece := range w.continuation[prev] {
			for to := range w.continuation[prev][piece] {
				w.continuation[prev][piece][to] /= 2
			}
		}
	}
}

func abs32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}
