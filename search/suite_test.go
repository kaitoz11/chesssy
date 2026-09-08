// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package search

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/kaitoz11/chesssy/engine"
)

// The nine position bench in the uci package is the engine's public speed figure. This
// one is wider, for judging changes to the search itself.
//
// A change to move ordering or pruning does not make a node cheaper, it changes how
// many nodes are needed, and nine positions is too few to tell a real improvement from
// luck in how ties happen to break. Forty positions taken from self-play across eight
// openings give a node total stable enough to compare: single threaded, it is exactly
// reproducible, so any difference in it is caused by the change and nothing else.
//
//	go test ./search -run XXX -bench BenchmarkSuite -benchtime 1x
//
// Report both numbers. Fewer nodes in the same time is a better search; the same nodes
// in less time is a faster one.

// suitePositions returns the measurement positions.
func suitePositions(tb testing.TB) []string {
	tb.Helper()

	f, err := os.Open("testdata/positions.fen")
	if err != nil {
		tb.Fatalf("measurement suite: %v", err)
	}
	defer f.Close()

	var positions []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			positions = append(positions, line)
		}
	}
	if err := scanner.Err(); err != nil {
		tb.Fatal(err)
	}
	if len(positions) == 0 {
		tb.Fatal("measurement suite is empty")
	}
	return positions
}

func BenchmarkSuite(b *testing.B) {
	positions := suitePositions(b)

	for _, depth := range []int{8, 11} {
		b.Run(fmt.Sprintf("depth-%d", depth), func(b *testing.B) {
			searcher := NewSearcher(64)
			var nodes uint64

			for i := 0; i < b.N; i++ {
				for _, fen := range positions {
					p, err := engine.NewPositionFromFEN(fen)
					if err != nil {
						b.Fatalf("%s: %v", fen, err)
					}
					// Each position starts from a cleared table, so that the node
					// count does not depend on the order they are searched in.
					searcher.NewGame()
					nodes += searcher.Search(p, Limits{Depth: depth}).Nodes
				}
			}

			b.ReportMetric(float64(nodes)/float64(b.N), "nodes")
			b.ReportMetric(float64(nodes)/b.Elapsed().Seconds(), "nodes/s")
		})
	}
}

// TestSuiteIsReproducible guards the measurement itself: if the same search of the same
// position ever returns a different node count, every comparison made with it is
// meaningless.
func TestSuiteIsReproducible(t *testing.T) {
	positions := suitePositions(t)[:5]
	searcher := NewSearcher(16)

	var first []uint64
	for run := range 2 {
		for i, fen := range positions {
			p, err := engine.NewPositionFromFEN(fen)
			if err != nil {
				t.Fatal(err)
			}
			searcher.NewGame()
			nodes := searcher.Search(p, Limits{Depth: 8}).Nodes

			if run == 0 {
				first = append(first, nodes)
				continue
			}
			if nodes != first[i] {
				t.Errorf("position %d searched %d nodes, then %d", i, first[i], nodes)
			}
		}
	}
}
