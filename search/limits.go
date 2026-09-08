package search

import (
	"time"

	"github.com/kaitoz11/chesssy/engine"
)

// Limits says when the search must stop. The zero value means "search until Stop is
// called". Whichever limit is reached first wins.
type Limits struct {
	// Depth is the deepest iteration to complete, 0 for no depth limit.
	Depth int
	// Nodes is a node budget, 0 for no node limit.
	Nodes uint64

	// MoveTime is a fixed allowance for this move. It takes precedence over the
	// clock fields below.
	MoveTime time.Duration
	// Time and Increment are the game clock: time left and the increment added
	// per move, per side.
	Time      [engine.ColorCount]time.Duration
	Increment [engine.ColorCount]time.Duration
	// MovesToGo is the number of moves left in the current time period, 0 when the
	// clock is not periodic.
	MovesToGo int

	// Infinite ignores the clock entirely; only Stop ends such a search.
	Infinite bool
}

// Time management works with two deadlines.
//
// The soft limit is the budget for the move. It is checked between iterations: an
// iteration typically costs several times the previous one, so once the soft limit
// is close there is no point starting another.
//
// The hard limit is the point of no return, checked inside the search. Hitting it
// abandons the current iteration and falls back on the last completed one.
const (
	// moveOverhead is held back for the time the host takes to relay a move.
	moveOverhead = 25 * time.Millisecond

	// assumedMovesToGo is how many more moves the game is assumed to last when the
	// host does not say.
	assumedMovesToGo = 30

	// incrementFraction is how much of the increment is spent on this move rather
	// than saved.
	incrementNumerator   = 3
	incrementDenominator = 4

	// The soft limit never exceeds softLimitCap of the remaining clock, and the
	// hard limit never exceeds hardLimitCap, so that a single move cannot spend
	// the whole game.
	softLimitCapNumerator   = 2
	softLimitCapDenominator = 3
	hardLimitCapNumerator   = 4
	hardLimitCapDenominator = 5

	// hardLimitMultiple is how far past the budget a single iteration may run
	// before it is abandoned.
	hardLimitMultiple = 4

	// nextIterationFraction is the share of the soft limit that must remain unspent
	// for another iteration to be worth starting.
	nextIterationNumerator   = 2
	nextIterationDenominator = 3
)

// setTimeLimits turns the limits into the two deadlines the search checks. Both are
// left at zero when there is no clock, which means "no deadline".
func (w *worker) setTimeLimits(side engine.Color) {
	w.softLimit, w.hardLimit = 0, 0

	switch {
	case w.limits.Infinite:
		return

	case w.limits.MoveTime > 0:
		w.hardLimit = max(w.limits.MoveTime-moveOverhead, time.Millisecond)
		w.softLimit = w.hardLimit

	case w.limits.Time[side] > 0:
		remaining := max(w.limits.Time[side]-moveOverhead, time.Millisecond)
		movesToGo := w.limits.MovesToGo
		if movesToGo <= 0 {
			movesToGo = assumedMovesToGo
		}

		budget := remaining/time.Duration(movesToGo) +
			w.limits.Increment[side]*incrementNumerator/incrementDenominator

		w.softLimit = min(budget, remaining*softLimitCapNumerator/softLimitCapDenominator)
		w.hardLimit = min(budget*hardLimitMultiple, remaining*hardLimitCapNumerator/hardLimitCapDenominator)
	}
}

// checkLimits publishes this worker's node count and, on the main thread, aborts the
// search when the node budget or the hard deadline is spent. It is called every few
// thousand nodes rather than every node, since reading the clock is not free.
//
// Only the main thread watches the clock. A helper stops when the main thread does,
// which keeps the decision in one place and saves every helper a clock read.
func (w *worker) checkLimits() {
	w.published.Store(w.nodes)
	if !w.main {
		return
	}
	if w.limits.Nodes > 0 && w.nodeCount() >= w.limits.Nodes {
		w.abort.Store(true)
		return
	}
	if w.hardLimit > 0 && time.Since(w.start) >= w.hardLimit {
		w.abort.Store(true)
	}
}

// outOfTimeForNextIteration reports whether too little of the budget is left for
// another iteration to finish.
func (w *worker) outOfTimeForNextIteration() bool {
	if w.softLimit <= 0 {
		return false
	}
	spent := time.Since(w.start)
	return spent >= w.softLimit*nextIterationNumerator/nextIterationDenominator
}
