// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package search

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/kaitoz11/chesssy/engine"
)

// Searcher is the public face of the search: it owns the transposition table, the
// worker threads and their accumulated knowledge.
//
// Reuse one Searcher for a whole game, since what it learns is what makes later moves
// cheaper. One Searcher runs one search at a time; Stop may be called from another
// goroutine while a search is in progress.
//
// # Threads
//
// With more than one thread the search is "lazy SMP": every thread searches the same
// position independently, sharing only the transposition table. There is no work
// splitting and no synchronisation on the search itself. Threads start at staggered
// depths and, because each keeps its own ordering heuristics, quickly explore
// different parts of the tree; whatever one thread proves, the others find waiting for
// them in the table. It is a crude form of parallelism that works remarkably well, and
// it costs nothing in the single threaded case.
type Searcher struct {
	// OnInfo, when set, is called once per completed iteration of the main thread,
	// from the goroutine running that thread.
	OnInfo func(Info)

	tt      *Table
	workers []*worker
	abort   atomic.Bool
}

// NewSearcher returns a single threaded searcher with a transposition table of about
// sizeMB megabytes.
func NewSearcher(hashSizeMB int) *Searcher {
	s := &Searcher{tt: NewTable(hashSizeMB)}
	s.SetThreads(1)
	return s
}

// SetThreads sets how many threads a search uses. One means the search runs entirely
// on the calling goroutine and is deterministic; more threads search faster but make
// the node count, and occasionally the move, depend on timing.
func (s *Searcher) SetThreads(n int) {
	n = max(n, 1)
	if n == len(s.workers) {
		return
	}
	// Growing keeps what the existing workers have learned.
	for len(s.workers) < n {
		s.workers = append(s.workers, newWorker(len(s.workers), s.tt, &s.abort))
	}
	s.workers = s.workers[:n]
}

// Threads returns how many threads a search will use.
func (s *Searcher) Threads() int { return len(s.workers) }

// SetHashSize replaces the transposition table, discarding its contents.
func (s *Searcher) SetHashSize(sizeMB int) {
	s.tt = NewTable(sizeMB)
	for _, w := range s.workers {
		w.tt = s.tt
	}
}

// HashFull reports transposition table occupancy in per mille, as UCI expects.
func (s *Searcher) HashFull() int { return s.tt.Hashfull() }

// NewGame forgets everything learned from earlier positions. Call it when the next
// search is unrelated to the last, so that stale scores and ordering cannot mislead
// it.
func (s *Searcher) NewGame() {
	s.tt.Clear()
	for _, w := range s.workers {
		w.forget()
	}
}

// Stop asks a running search to return as soon as it can. It is safe to call from
// another goroutine, and safe to call when no search is running.
func (s *Searcher) Stop() { s.abort.Store(true) }

// Search runs iterative deepening on pos and returns the last completed iteration of
// the main thread. It leaves pos exactly as it was passed in.
func (s *Searcher) Search(pos *engine.Position, limits Limits) Info {
	s.begin(pos, limits)
	return s.run()
}

// SearchAsync starts a search on a new goroutine and calls onDone with the result.
//
// The searcher is reset before SearchAsync returns, so a Stop issued afterwards is
// always seen by this search. That is the difference from calling Search in a
// goroutine, where the reset would race with the Stop.
func (s *Searcher) SearchAsync(pos *engine.Position, limits Limits, onDone func(Info)) {
	s.begin(pos, limits)
	go func() {
		info := s.run()
		if onDone != nil {
			onDone(info)
		}
	}()
}

// begin prepares every worker for a new search.
func (s *Searcher) begin(pos *engine.Position, limits Limits) {
	s.abort.Store(false)
	s.tt.NewSearch()

	start := time.Now()
	for _, w := range s.workers {
		w.onInfo = nil
		w.totalNodes = s.Nodes
		if w.main {
			w.onInfo = s.OnInfo
		}
		// Each worker searches its own copy, so that they never write to the same
		// board.
		w.begin(pos.Clone(), limits, start)
	}
}

// run searches with every worker and returns the main thread's result.
func (s *Searcher) run() Info {
	var helpers sync.WaitGroup
	for _, w := range s.workers[1:] {
		helpers.Add(1)
		go func(w *worker) {
			defer helpers.Done()
			w.iterate()
		}(w)
	}

	info := s.workers[0].iterate()

	// The main thread has its answer, so the helpers have nothing left to contribute.
	s.abort.Store(true)
	helpers.Wait()

	info.Nodes = s.Nodes()
	return info
}

// Nodes returns the nodes searched by every worker in the current or last search.
func (s *Searcher) Nodes() uint64 {
	var total uint64
	for _, w := range s.workers {
		total += w.published.Load()
	}
	return total
}
