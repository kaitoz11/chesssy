// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package search

import (
	"slices"
	"sync/atomic"
	"testing"

	"github.com/kaitoz11/chesssy/engine"
)

// pickerPositions cover the cases move ordering has to get right: quiet openings,
// tactical middlegames with captures on both sides, promotions, and being in check.
var pickerPositions = []string{
	engine.StartFEN,
	"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
	"r2q1rk1/pP1p2pp/Q4n2/bbp1p3/Np6/1B3NBn/pPPP1PPP/R3K2R b KQ - 0 1",
	"rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
	"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
	"4k3/8/8/8/8/8/4R3/4K2R b - - 0 1", // in check
	"n1n5/PPPk4/8/8/8/8/4Kppp/5N1N b - - 0 1",
}

// TestPickerYieldsEveryMoveOnce is the correctness contract of the picker: staging
// and partitioning reorder the list in place, and a bug there would silently drop or
// duplicate a move, which no perft or search test would catch.
func TestPickerYieldsEveryMoveOnce(t *testing.T) {
	for _, fen := range pickerPositions {
		t.Run(fen, func(t *testing.T) {
			p, err := engine.NewPositionFromFEN(fen)
			if err != nil {
				t.Fatal(err)
			}
			w := testWorker(p)

			var list engine.MoveList
			p.GenerateMoves(&list)
			want := slices.Clone(list.Moves())

			// Try it both with and without a table move, since that changes where
			// the partition starts.
			for _, ttMove := range []engine.Move{engine.NoMove, want[len(want)-1]} {
				got := drain(w, &list, ttMove)

				if len(got) != len(want) {
					t.Fatalf("picker yielded %d moves, want %d", len(got), len(want))
				}
				sortedGot, sortedWant := slices.Clone(got), slices.Clone(want)
				slices.Sort(sortedGot)
				slices.Sort(sortedWant)
				if !slices.Equal(sortedGot, sortedWant) {
					t.Errorf("picker yielded a different set of moves\ngot  %v\nwant %v", got, want)
				}
				if ttMove != engine.NoMove && got[0] != ttMove {
					t.Errorf("table move %s was handed out at position %d, want first",
						ttMove, slices.Index(got, ttMove))
				}
			}
		})
	}
}

// TestPickerOrdersByClass checks the order the stages are meant to produce: winning
// captures and queen promotions, then quiet moves, then losing captures.
func TestPickerOrdersByClass(t *testing.T) {
	for _, fen := range pickerPositions {
		t.Run(fen, func(t *testing.T) {
			p, err := engine.NewPositionFromFEN(fen)
			if err != nil {
				t.Fatal(err)
			}
			w := testWorker(p)

			var list engine.MoveList
			p.GenerateMoves(&list)
			got := drain(w, &list, engine.NoMove)

			// class 0 = winning capture or queen promotion, 1 = quiet, 2 = losing
			// capture. The classes must appear in that order.
			class := func(m engine.Move) int {
				queenPromo := m.IsPromotion() && m.Promotion() == engine.Queen
				switch {
				case queenPromo || (p.IsCapture(m) && p.SeeGE(m, 0)):
					return 0
				case p.IsCapture(m):
					return 2
				default:
					return 1
				}
			}
			highest := 0
			for i, m := range got {
				c := class(m)
				if c < highest {
					t.Errorf("move %d (%s) is class %d after class %d: %v", i, m, c, highest, got)
					break
				}
				highest = c
			}
		})
	}
}

// testWorker returns a worker searching p, for tests that exercise the pieces of the
// search below the Searcher API.
func testWorker(p *engine.Position) *worker {
	var abort atomic.Bool
	w := newWorker(0, NewTable(1), &abort)
	w.pos = p
	return w
}

// drain returns every move the picker hands out.
func drain(w *worker, list *engine.MoveList, ttMove engine.Move) []engine.Move {
	var picker movePicker
	picker.init(w, list, ttMove, 0)

	var got []engine.Move
	for m := picker.next(); m != engine.NoMove; m = picker.next() {
		got = append(got, m)
	}
	return got
}
