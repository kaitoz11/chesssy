package search

import (
	"testing"

	"github.com/kaitoz11/chesssy/engine"
)

func TestTranspositionTableRoundTrip(t *testing.T) {
	tt := NewTable(1)
	key := uint64(0xDEADBEEFCAFEBABE)
	want := Entry{
		Move:       engine.NewMove(engine.E2, engine.E4),
		Score:      123,
		StaticEval: 45,
		Depth:      7,
		Bound:      BoundExact,
	}
	_, _, slot := tt.Probe(key)
	tt.Store(slot, key, want)

	got, found, _ := tt.Probe(key)
	if !found {
		t.Fatal("stored entry not found")
	}
	if got != want {
		t.Errorf("Probe() = %+v, want %+v", got, want)
	}
	if _, found, _ := tt.Probe(key + 1<<32); found {
		t.Error("probe with a different key returned a hit")
	}

	tt.Clear()
	if _, found, _ := tt.Probe(key); found {
		t.Error("entry survived Clear")
	}
}

func TestTranspositionTableScoreAdjustment(t *testing.T) {
	// Mate scores must be stored relative to the mating position so they stay
	// correct when the same entry is used at a different ply.
	for _, ply := range []int{0, 3, 17} {
		for _, score := range []int{mateScore - 5, -mateScore + 5, 42, -42} {
			stored := scoreToTT(score, ply)
			if got := scoreFromTT(stored, ply); got != score {
				t.Errorf("ply %d score %d: round trip gave %d", ply, score, got)
			}
		}
	}
}

func TestHashFullGrows(t *testing.T) {
	s := NewSearcher(1)
	if got := s.HashFull(); got != 0 {
		t.Errorf("fresh table reports %d per mille full", got)
	}
	s.Search(engine.NewPosition(), Limits{Depth: 8})
	if got := s.HashFull(); got == 0 {
		t.Error("table still reports empty after a search")
	}
}
