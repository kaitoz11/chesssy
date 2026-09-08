package search

import (
	"sync/atomic"
	"time"

	"github.com/kaitoz11/chesssy/engine"
)

// A worker is one search thread. It owns everything the search writes to except the
// transposition table, which is shared: its own copy of the position, its own move
// buffers, and its own ordering heuristics. Nothing in a worker is touched by another
// goroutine, so the search itself needs no synchronisation at all.
//
// All of it is preallocated, so a search does no allocation once it is running.
type worker struct {
	id  int
	tt  *Table
	pos *engine.Position

	// main marks the worker whose result is reported. It is the only one that
	// watches the clock and the only one that publishes progress.
	main   bool
	onInfo func(Info)
	// totalNodes reports the nodes searched by every worker, for reporting.
	totalNodes func() uint64
	// abort is shared: whoever decides the search is over sets it, and every worker
	// notices within a few thousand nodes.
	abort *atomic.Bool

	pawns *pawnCache

	// Ordering heuristics, kept between moves. See history.go.
	history [engine.PieceCount][engine.SquareCount]int32
	// continuation is history conditioned on the move that was just played, indexed by
	// where that move landed. See history.go.
	continuation [engine.SquareCount][engine.PieceCount][engine.SquareCount]int32
	counters     [engine.NoPiece + 1][engine.SquareCount]engine.Move
	killers      [MaxPly + 2][killerSlots]engine.Move

	// Per-ply scratch space. Indexing by ply rather than recursing with locals keeps
	// the buffers out of the stack frame and reused across nodes.
	lists     [MaxPly + 2]engine.MoveList
	pickers   [MaxPly + 2]movePicker
	moveStack [MaxPly + 2]engine.Move // the move played into each ply
	evalStack [MaxPly + 2]int         // the static evaluation at each ply

	// The principal variation, collected as a triangular table: pv[ply] holds the
	// best line found from that ply onwards.
	pv       [MaxPly + 2][MaxPly + 2]engine.Move
	pvLength [MaxPly + 2]int

	nodes    uint64
	selDepth int
	rootBest engine.Move

	// published is nodes, copied out at every limit check so that the reporting
	// thread can read a recent count without a data race on every node.
	published atomic.Uint64

	limits    Limits
	start     time.Time
	softLimit time.Duration
	hardLimit time.Duration
}

func newWorker(id int, tt *Table, abort *atomic.Bool) *worker {
	return &worker{
		id:    id,
		tt:    tt,
		abort: abort,
		main:  id == 0,
		pawns: &pawnCache{},
	}
}

// begin resets the per-search state. It runs before the search proper, and on the
// caller's goroutine, so that a Stop arriving afterwards is never lost.
func (w *worker) begin(pos *engine.Position, limits Limits, start time.Time) {
	w.pos = pos
	w.limits = limits
	w.nodes = 0
	w.published.Store(0)
	w.selDepth = 0
	w.rootBest = engine.NoMove
	w.start = start
	w.setTimeLimits(pos.SideToMove())

	w.killers = [MaxPly + 2][killerSlots]engine.Move{}
	w.moveStack = [MaxPly + 2]engine.Move{}
	for i := range w.evalStack {
		w.evalStack[i] = noEval
	}
	w.decayHistory()
}

// forget clears everything learned from earlier positions.
func (w *worker) forget() {
	w.history = [engine.PieceCount][engine.SquareCount]int32{}
	w.continuation = [engine.SquareCount][engine.PieceCount][engine.SquareCount]int32{}
	w.counters = [engine.NoPiece + 1][engine.SquareCount]engine.Move{}
	w.killers = [MaxPly + 2][killerSlots]engine.Move{}
}

func (w *worker) stopped() bool { return w.abort.Load() }

// evaluate scores the position statically, reusing this worker's pawn cache.
func (w *worker) evaluate(p *engine.Position) int { return evaluate(p, w.pawns) }

// nodeCount returns the nodes to report: every worker's when the search is shared,
// this worker's when it is running alone.
func (w *worker) nodeCount() uint64 {
	if w.totalNodes != nil {
		return w.totalNodes()
	}
	return w.nodes
}

// previousMove returns the move played into the node at ply, or engine.NoMove at the
// root or after a null move.
func (w *worker) previousMove(ply int) engine.Move {
	if ply == 0 {
		return engine.NoMove
	}
	return w.moveStack[ply-1]
}

// hasNonPawnMaterial reports whether c has a piece other than pawns and the king.
// Null move pruning needs it: with only pawns left, passing can be the best move
// available, and pruning on it would be wrong.
func (w *worker) hasNonPawnMaterial(c engine.Color) bool {
	p := w.pos
	return p.ColorBB(c) & ^p.TypeBB(engine.Pawn) & ^p.TypeBB(engine.King) != 0
}

// Helper threads follow a pattern of depths to skip, so that at any moment they are
// spread across different depths rather than all repeating the main thread's. A helper
// that is a ply or two ahead fills the transposition table with answers the main thread
// then finds waiting for it; one that duplicates the main thread's depth contributes
// almost nothing.
//
// The pattern is the one Stockfish uses: paired stride and phase tables, indexed by
// thread, which give each helper a different rhythm of searched and skipped depths.
var (
	skipStride = [20]int{1, 1, 2, 2, 2, 3, 3, 3, 4, 4, 4, 5, 5, 5, 5, 5, 6, 6, 6, 6}
	skipPhase  = [20]int{0, 1, 0, 1, 2, 0, 1, 2, 3, 0, 1, 0, 1, 2, 3, 4, 0, 1, 2, 3}
)

// startDepth is the iteration a worker begins at.
func (w *worker) startDepth() int {
	if w.main {
		return 1
	}
	return 1 + w.id%3
}

// skipDepth reports whether this worker should pass over an iteration. The main thread
// never skips: its results are the ones reported.
func (w *worker) skipDepth(depth int) bool {
	if w.main {
		return false
	}
	i := (w.id - 1) % len(skipStride)
	return ((depth+skipPhase[i])/skipStride[i])%2 == 1
}
