// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package search

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/kaitoz11/chesssy/engine"
)

// Node counts alone cannot say whether a search change was an improvement: pruning more
// aggressively always searches fewer nodes, and past some point it does so by missing
// the best move. This test is the counterweight.
//
// testdata/bestmoves.tsv holds, for each position of the measurement suite, the move a
// three second search on every core chose, at a mean depth of about twenty. Agreement is
// how often a shallow search picks that same move. It is a proxy, not a rating: a
// disagreement can be a transposition of equal value, and the deep search is the same
// engine and so shares its blind spots. But a change that cuts nodes while agreement
// falls is pruning too much, and that is exactly what the number is for.
//
//	go test ./search -run TestMoveAgreement -v

// agreementDepth is shallow enough to run in seconds and deep enough that a well ordered
// search usually finds the right move.
const agreementDepth = 9

// minAgreement is a floor, not a target. Half is a normal figure here: a nine ply search
// and a twenty ply search disagree often, and many of those disagreements are moves of
// equal value. The floor sits well under the measured level so that the test fails on a
// real regression rather than on ordinary variation.
const minAgreement = 35

type groundTruth struct {
	fen  string
	best engine.Move
}

func loadGroundTruth(tb testing.TB) []groundTruth {
	tb.Helper()

	f, err := os.Open("testdata/bestmoves.tsv")
	if err != nil {
		tb.Fatalf("ground truth: %v", err)
	}
	defer f.Close()

	var truth []groundTruth
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), "\t")
		if len(fields) < 2 {
			continue
		}
		p, err := engine.NewPositionFromFEN(fields[0])
		if err != nil {
			tb.Fatalf("ground truth position %q: %v", fields[0], err)
		}
		best, err := p.ParseMove(fields[1])
		if err != nil {
			tb.Fatalf("ground truth move %q: %v", fields[1], err)
		}
		truth = append(truth, groundTruth{fen: fields[0], best: best})
	}
	if err := scanner.Err(); err != nil {
		tb.Fatal(err)
	}
	if len(truth) == 0 {
		tb.Fatal("ground truth is empty")
	}
	return truth
}

func TestMoveAgreement(t *testing.T) {
	truth := loadGroundTruth(t)
	searcher := NewSearcher(64)

	agreed, nodes := 0, uint64(0)
	for _, want := range truth {
		p, err := engine.NewPositionFromFEN(want.fen)
		if err != nil {
			t.Fatal(err)
		}
		searcher.NewGame()
		info := searcher.Search(p, Limits{Depth: agreementDepth})
		nodes += info.Nodes

		if info.BestMove() == want.best {
			agreed++
			continue
		}
		t.Logf("disagree on %s: depth %d chose %s, deep search chose %s",
			want.fen, agreementDepth, info.BestMove(), want.best)
	}

	percent := agreed * 100 / len(truth)
	t.Logf("agreement %d%% (%d of %d) at depth %d, %d nodes",
		percent, agreed, len(truth), agreementDepth, nodes)

	if percent < minAgreement {
		t.Errorf("agreement fell to %d%%, below the %d%% floor: the search is missing moves it used to find",
			percent, minAgreement)
	}
}
