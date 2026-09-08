package search

import (
	"testing"
	"time"

	"github.com/kaitoz11/chesssy/engine"
)

// With more than one thread the search is no longer deterministic, so these tests
// check the properties that must hold however the threads interleave: the move played
// is legal, forced mates are still found, the position is left untouched, and stopping
// works. Run them under the race detector; that is where a shared-state mistake in the
// transposition table or the worker split would surface.

const testThreads = 4

func TestThreadedSearchFindsForcedMates(t *testing.T) {
	s := NewSearcher(16)
	s.SetThreads(testThreads)
	if got := s.Threads(); got != testThreads {
		t.Fatalf("Threads() = %d, want %d", got, testThreads)
	}

	checked := 0
	for _, fen := range matePositions {
		p, err := engine.NewPositionFromFEN(fen)
		if err != nil {
			t.Fatalf("%s: %v", fen, err)
		}
		plies := forcedMatePlies(p, 5)
		if plies == 0 {
			continue
		}
		checked++

		info := s.Search(p, Limits{Depth: 12})
		if info.Mate != (plies+1)/2 {
			t.Errorf("%s: mate score %d, want %d (pv %v)", fen, info.Mate, (plies+1)/2, info.PV)
		}
		if got := p.FEN(); got != fen {
			t.Errorf("%s: search left the position as %s", fen, got)
		}
	}
	if checked < 4 {
		t.Fatalf("only %d positions had a forced mate; the suite is not testing much", checked)
	}
}

func TestThreadedSearchReturnsLegalMoves(t *testing.T) {
	s := NewSearcher(16)
	s.SetThreads(testThreads)

	for _, fen := range []string{
		engine.StartFEN,
		"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		"r2q1rk1/pP1p2pp/Q4n2/bbp1p3/Np6/1B3NBn/pPPP1PPP/R3K2R b KQ - 0 1",
	} {
		p, err := engine.NewPositionFromFEN(fen)
		if err != nil {
			t.Fatal(err)
		}
		info := s.Search(p, Limits{Depth: 8})

		var list engine.MoveList
		p.GenerateMoves(&list)
		if !list.Contains(info.BestMove()) {
			t.Errorf("%s: best move %s is not legal", fen, info.BestMove())
		}
		if info.Nodes == 0 {
			t.Errorf("%s: reported no nodes", fen)
		}
	}
}

func TestThreadedSearchAddsThroughput(t *testing.T) {
	p := engine.NewPosition()
	const budget = 250 * time.Millisecond

	single := NewSearcher(16)
	solo := single.Search(p, Limits{MoveTime: budget})

	many := NewSearcher(16)
	many.SetThreads(testThreads)
	shared := many.Search(p, Limits{MoveTime: budget})

	// Given the same wall time, more threads have to get through more nodes. Counting
	// per depth instead would prove nothing: helpers skip depths on purpose, so at a
	// fixed depth they can legitimately search less than one thread does.
	//
	// Depth is not asserted for the same reason, and because under the race detector the
	// instrumentation and the contention between threads can leave four of them a ply
	// behind one, which says nothing about the search.
	if shared.Nodes <= solo.Nodes {
		t.Errorf("in %v, %d threads searched %d nodes, not more than one thread's %d",
			budget, testThreads, shared.Nodes, solo.Nodes)
	}
}

func TestThreadedStopEndsSearch(t *testing.T) {
	p := engine.NewPosition()
	s := NewSearcher(16)
	s.SetThreads(testThreads)

	done := make(chan Info, 1)
	go func() { done <- s.Search(p, Limits{Infinite: true}) }()

	time.Sleep(150 * time.Millisecond)
	s.Stop()

	select {
	case info := <-done:
		if info.BestMove() == engine.NoMove {
			t.Error("stopped search returned no move")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("threaded search did not stop")
	}
}

func TestSetThreadsIsIdempotentAndClamped(t *testing.T) {
	s := NewSearcher(1)
	if got := s.Threads(); got != 1 {
		t.Errorf("a new searcher has %d threads, want 1", got)
	}
	for _, n := range []int{0, -3, 1} {
		s.SetThreads(n)
		if got := s.Threads(); got != 1 {
			t.Errorf("SetThreads(%d) gave %d threads, want 1", n, got)
		}
	}
	s.SetThreads(3)
	if got := s.Threads(); got != 3 {
		t.Errorf("SetThreads(3) gave %d threads", got)
	}
	// Shrinking and growing again must leave a usable searcher.
	s.SetThreads(1)
	s.SetThreads(2)
	if info := s.Search(engine.NewPosition(), Limits{Depth: 6}); info.BestMove() == engine.NoMove {
		t.Error("searcher stopped working after resizing the thread pool")
	}
}

func TestThreadedSelfPlayStaysLegal(t *testing.T) {
	p := engine.NewPosition()
	s := NewSearcher(16)
	s.SetThreads(testThreads)
	s.NewGame()

	for ply := 0; ply < 60; ply++ {
		if !p.HasLegalMoves() || p.IsFiftyMoveDraw() || p.HasInsufficientMaterial() {
			break
		}
		info := s.Search(p, Limits{Depth: 6})
		m := info.BestMove()

		var list engine.MoveList
		p.GenerateMoves(&list)
		if !list.Contains(m) {
			t.Fatalf("ply %d: illegal move %s in %s", ply, m, p.FEN())
		}
		p.MakeMove(m)
	}
}
