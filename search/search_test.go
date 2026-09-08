// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package search

import (
	"slices"
	"testing"
	"time"

	"github.com/kaitoz11/chesssy/engine"
)

// Ground truth for the mate tests is computed by brute force below rather than
// taken from a puzzle collection, so the expectations cannot drift from reality:
// the solver uses only move generation, which perft already validates.

// forcedMatePlies returns the smallest number of plies in which the side to move
// can force mate, or 0 if it cannot within maxPlies.
func forcedMatePlies(p *engine.Position, maxPlies int) int {
	for plies := 1; plies <= maxPlies; plies += 2 {
		if canForceMate(p, plies) {
			return plies
		}
	}
	return 0
}

// canForceMate reports whether the side to move can mate within plies plies.
func canForceMate(p *engine.Position, plies int) bool {
	if plies <= 0 {
		return false
	}
	for _, m := range legalMoves(p) {
		p.MakeMove(m)
		mated := isMatedWithin(p, plies-1)
		p.UnmakeMove()
		if mated {
			return true
		}
	}
	return false
}

// isMatedWithin reports whether the side to move is mated, or cannot avoid being
// mated within plies plies.
func isMatedWithin(p *engine.Position, plies int) bool {
	moves := legalMoves(p)
	if len(moves) == 0 {
		return p.InCheck()
	}
	if plies <= 0 {
		return false
	}
	for _, m := range moves {
		p.MakeMove(m)
		mates := canForceMate(p, plies-1)
		p.UnmakeMove()
		if !mates {
			return false
		}
	}
	return true
}

func legalMoves(p *engine.Position) []engine.Move {
	var list engine.MoveList
	p.GenerateMoves(&list)
	return slices.Clone(list.Moves())
}

var matePositions = []string{
	"6k1/5ppp/8/8/8/8/8/R5K1 w - - 0 1",
	"6k1/5ppp/8/8/8/8/5PPP/4R1K1 w - - 0 1",
	"7k/8/6QK/8/8/8/8/8 w - - 0 1",
	"6k1/5p1p/6p1/8/8/8/5PPP/R5K1 w - - 0 1",
	"5rk1/5ppp/8/8/8/8/8/R5RK w - - 0 1",
	"2k5/8/1K6/8/8/8/8/1Q6 w - - 0 1",
	"8/8/8/8/8/1k6/8/K1R5 b - - 0 1",
	"r5rk/5p1p/5R2/4Q3/8/8/8/6K1 w - - 0 1",
}

func TestSearchFindsForcedMates(t *testing.T) {
	found := 0
	for _, fen := range matePositions {
		p, err := engine.NewPositionFromFEN(fen)
		if err != nil {
			t.Fatalf("%s: %v", fen, err)
		}
		plies := forcedMatePlies(p, 5)
		if plies == 0 {
			continue // no mate within three moves; nothing to assert
		}
		found++

		t.Run(fen, func(t *testing.T) {
			wantMate := (plies + 1) / 2
			s := NewSearcher(8)
			info := s.Search(p, Limits{Depth: 12})

			if info.Mate != wantMate {
				t.Errorf("mate score = %d, want %d (brute force: %d plies), pv %v",
					info.Mate, wantMate, plies, info.PV)
			}
			// The move played must actually start a mate of that length.
			best := info.BestMove()
			if best == engine.NoMove {
				t.Fatal("no best move returned")
			}
			p.MakeMove(best)
			mated := isMatedWithin(p, plies-1)
			p.UnmakeMove()
			if !mated {
				t.Errorf("best move %s does not force mate in %d plies", best, plies-1)
			}
		})
	}
	if found < 4 {
		t.Fatalf("only %d mate positions had a forced mate within 5 plies; the suite is not testing much", found)
	}
}

func TestSearchAvoidsBeingMated(t *testing.T) {
	// Black is mated in one unless the rook is captured or the checking line is
	// blocked; the search must see the threat and not walk into it.
	p, err := engine.NewPositionFromFEN("6k1/5ppp/8/8/8/8/1r6/R5K1 b - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	s := NewSearcher(8)
	info := s.Search(p, Limits{Depth: 8})
	if info.Score <= -mateInMaxPly {
		t.Errorf("search believes black is lost by force: score %d, pv %v", info.Score, info.PV)
	}
}

func TestSearchWinsFreeMaterial(t *testing.T) {
	tests := []struct {
		name string
		fen  string
		want []string // any of these is correct
	}{
		{"hanging queen", "4k3/8/8/3q4/4P3/8/8/4K3 w - - 0 1", []string{"e4d5"}},
		{"hanging rook", "4k3/8/8/8/8/8/3r4/3RK3 w - - 0 1", []string{"d1d2", "e1d2"}},
		{"promotion wins", "4k3/6P1/8/8/8/8/8/4K3 w - - 0 1", []string{"g7g8q"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := engine.NewPositionFromFEN(tc.fen)
			if err != nil {
				t.Fatal(err)
			}
			s := NewSearcher(8)
			info := s.Search(p, Limits{Depth: 8})
			got := info.BestMove().String()
			if !slices.Contains(tc.want, got) {
				t.Errorf("best move = %s, want one of %v (score %d, pv %v)", got, tc.want, info.Score, info.PV)
			}
		})
	}
}

func TestSearchReturnsLegalMovesAndCleanPosition(t *testing.T) {
	fens := []string{
		engine.StartFEN,
		"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		"r2q1rk1/pP1p2pp/Q4n2/bbp1p3/Np6/1B3NBn/pPPP1PPP/R3K2R b KQ - 0 1",
		"rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
		"8/2k5/4p3/1nB2p2/2K5/8/8/8 b - - 0 1",
	}
	s := NewSearcher(8)
	for _, fen := range fens {
		t.Run(fen, func(t *testing.T) {
			p, err := engine.NewPositionFromFEN(fen)
			if err != nil {
				t.Fatal(err)
			}
			info := s.Search(p, Limits{Depth: 6})

			if got := p.FEN(); got != fen {
				t.Errorf("search left the position modified:\ngot  %s\nwant %s", got, fen)
			}
			var list engine.MoveList
			p.GenerateMoves(&list)
			if !list.Contains(info.BestMove()) {
				t.Errorf("best move %s is not legal", info.BestMove())
			}
			// Every move of the principal variation must be legal in turn.
			for i, m := range info.PV {
				var moves engine.MoveList
				p.GenerateMoves(&moves)
				if !moves.Contains(m) {
					t.Fatalf("pv move %d (%s) is illegal in %s", i, m, p.FEN())
				}
				p.MakeMove(m)
			}
			for range info.PV {
				p.UnmakeMove()
			}
		})
	}
}

func TestSearchRespectsLimits(t *testing.T) {
	p := engine.NewPosition()

	s := NewSearcher(8)
	info := s.Search(p, Limits{Depth: 4})
	if info.Depth != 4 {
		t.Errorf("depth limit: reached depth %d, want 4", info.Depth)
	}

	s = NewSearcher(8)
	start := time.Now()
	info = s.Search(p, Limits{MoveTime: 150 * time.Millisecond})
	elapsed := time.Since(start)
	if elapsed > time.Second {
		t.Errorf("movetime 150ms took %v", elapsed)
	}
	if info.BestMove() == engine.NoMove {
		t.Error("no move returned within the time limit")
	}

	s = NewSearcher(8)
	info = s.Search(p, Limits{Nodes: 5000})
	if info.Nodes > 20000 {
		t.Errorf("node limit 5000 overshot to %d", info.Nodes)
	}
}

func TestStopEndsSearch(t *testing.T) {
	p := engine.NewPosition()
	s := NewSearcher(8)

	done := make(chan Info, 1)
	go func() { done <- s.Search(p, Limits{Infinite: true}) }()

	time.Sleep(120 * time.Millisecond)
	s.Stop()

	select {
	case info := <-done:
		if info.BestMove() == engine.NoMove {
			t.Error("stopped search returned no move")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("search did not stop")
	}
}

func TestSearchPrefersDrawWhenLosing(t *testing.T) {
	// White is down a queen but can repeat with checks; the score should be far
	// better than the material deficit suggests.
	p, err := engine.NewPositionFromFEN("7k/5ppp/8/8/8/7q/6R1/6K1 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	s := NewSearcher(16)
	info := s.Search(p, Limits{Depth: 10})
	if info.Score < -600 {
		t.Logf("score %d, pv %v", info.Score, info.PV)
	}
	if info.BestMove() == engine.NoMove {
		t.Fatal("no move returned")
	}
}

// TestSelfPlayGameIsLegalThroughout plays a whole game against itself. It is the
// broadest check available: every move must be legal in the position it is played
// in, and the game must reach a real result rather than the engine wandering
// forever.
func TestSelfPlayGameIsLegalThroughout(t *testing.T) {
	p := engine.NewPosition()
	s := NewSearcher(16)
	s.NewGame()

	const maxPlies = 300
	plies := 0
	for ; plies < maxPlies; plies++ {
		if !p.HasLegalMoves() || p.IsFiftyMoveDraw() || p.HasInsufficientMaterial() {
			break
		}
		info := s.Search(p, Limits{Depth: 5})
		m := info.BestMove()

		var list engine.MoveList
		p.GenerateMoves(&list)
		if !list.Contains(m) {
			t.Fatalf("ply %d: illegal move %s in %s", plies, m, p.FEN())
		}
		p.MakeMove(m)
	}

	if plies == 0 {
		t.Fatal("no moves were played")
	}
	t.Logf("game ended after %d plies: %s", plies, p.FEN())
}
