package search

import "github.com/kaitoz11/chesssy/engine"

// Move ordering is the single biggest factor in alpha-beta efficiency: the sooner
// the best move is tried, the more of the tree the search can skip.
//
// movePicker hands a node's moves out in that order. It does two things to keep ordering
// cheap. The transposition table move goes first without scoring anything else, because
// it causes a cutoff often enough that scoring the rest would be wasted work. The
// remaining moves are then scored once and handed out by selection, so a node that cuts
// off early never pays for a full sort.
//
// Sorting the tail once, instead of scanning it per move, was measured and was slower on
// both counts: the extra array cost more in cache pressure than the comparisons it
// saved, and the change in how ties broke cost 12% more nodes.
type movePicker struct {
	worker *worker
	list   *engine.MoveList
	scores [engine.MaxMoves]int32
	ttMove engine.Move
	ply    int
	index  int
	// count is the move count. The selection loop reads it on every comparison, and
	// reaching through the list for it costs more than the comparison does.
	count   int
	ttFirst bool // the table move is parked at index 0 and needs no score
	scored  bool
}

// Ordering scores. The bands are far apart so that a move's class always outranks
// its refinement: any winning capture beats any killer, which beats any move
// ranked only by history.
const (
	scoreTTMove      = 1 << 24
	scoreQueenPromo  = 1<<22 + 1<<21
	scoreWinningCap  = 1 << 22
	scoreKiller      = 1 << 20
	scoreCounterMove = 1 << 19
	scoreLosingCap   = -(1 << 22)
)

// init prepares the picker for a freshly generated list.
func (mp *movePicker) init(w *worker, list *engine.MoveList, ttMove engine.Move, ply int) {
	mp.worker = w
	mp.list = list
	mp.ttMove = ttMove
	mp.ply = ply
	mp.index = 0
	mp.count = list.Len()
	mp.scored = false

	mp.ttFirst = false
	if ttMove != engine.NoMove {
		if i := list.IndexOf(ttMove); i >= 0 {
			list.Swap(0, i)
			mp.ttFirst = true
		}
	}
}

// next returns the next move to try, or engine.NoMove once the list is exhausted.
func (mp *movePicker) next() engine.Move {
	if mp.index >= mp.count {
		return engine.NoMove
	}
	if mp.index == 0 && mp.ttFirst {
		mp.index++
		return mp.list.At(0)
	}
	if !mp.scored {
		mp.scoreRemaining()
		mp.scored = true
	}

	m := mp.selectBest()
	mp.index++
	return m
}

// scoreRemaining scores the moves that have not been handed out yet. Skipping the
// ones already tried is what makes the table move free.
func (mp *movePicker) scoreRemaining() {
	w := mp.worker
	p := w.pos
	counter := w.counterMove(mp.ply)

	for i := mp.index; i < mp.count; i++ {
		m := mp.list.At(i)
		switch {
		case m == mp.ttMove:
			mp.scores[i] = scoreTTMove

		case m.IsPromotion() && m.Promotion() == engine.Queen:
			mp.scores[i] = scoreQueenPromo + int32(mvvLva(p, m))

		case p.IsCapture(m):
			// Winning and losing captures sit either side of the quiet moves, so
			// a losing capture is tried after the killers and the history.
			band := int32(scoreWinningCap)
			if !p.SeeGE(m, 0) {
				band = scoreLosingCap
			}
			mp.scores[i] = band + int32(mvvLva(p, m))

		case m == w.killers[mp.ply][0]:
			mp.scores[i] = scoreKiller + 1

		case m == w.killers[mp.ply][1]:
			mp.scores[i] = scoreKiller

		case m == counter:
			mp.scores[i] = scoreCounterMove

		default:
			mp.scores[i] = w.quietScore(mp.ply, p.PieceAt(m.From()), m)
		}
	}
}

// selectBest swaps the best remaining move into the current position and returns
// it.
func (mp *movePicker) selectBest() engine.Move {
	// Keeping the running best in locals lets the compiler drop the repeated
	// bounds checks and reloads.
	scores := mp.scores[:mp.count]
	best, bestScore := mp.index, scores[mp.index]
	for i := mp.index + 1; i < len(scores); i++ {
		if scores[i] > bestScore {
			best, bestScore = i, scores[i]
		}
	}
	if best != mp.index {
		mp.list.Swap(mp.index, best)
		scores[mp.index], scores[best] = scores[best], scores[mp.index]
	}
	return mp.list.At(mp.index)
}

// mvvLva ranks a capture by the value of its victim first and its attacker second,
// so that a pawn taking a queen is tried before a queen taking a pawn. The name
// stands for most valuable victim, least valuable attacker.
func mvvLva(p *engine.Position, m engine.Move) int {
	victim := engine.Pawn // en passant, whose destination square is empty
	if captured := p.PieceAt(m.To()); captured != engine.NoPiece {
		victim = captured.Type()
	}
	attacker := p.PieceAt(m.From()).Type()
	return engine.SeeValue(victim)*16 - engine.SeeValue(attacker)/16
}
