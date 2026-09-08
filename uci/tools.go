// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package uci

import (
	"strconv"
	"time"

	"github.com/kaitoz11/chesssy/engine"
	"github.com/kaitoz11/chesssy/search"
)

// Two development tools, exposed as commands so that they can be driven the same
// way as a search.
//
// perft counts the leaf nodes of the move tree, which is how move generation is
// verified: the counts for well-known positions are published. bench searches a
// fixed set of positions, which gives a speed figure that is comparable between
// builds.

// Defaults when the command is given without a depth.
const (
	defaultPerftDepth = 5
	defaultBenchDepth = 8
)

func (e *Engine) handlePerft(args []string) bool {
	e.runPerft(depthArg(args, defaultPerftDepth))
	return false
}

func (e *Engine) runPerft(depth int) {
	depth = max(depth, 1)

	start := time.Now()
	counts, total := e.position.PerftDivide(depth)
	elapsed := time.Since(start)

	e.printf("%s", engine.FormatDivide(counts, total))
	e.printf("depth %d, time %v, %.0f nodes/s\n",
		depth, elapsed.Round(time.Millisecond), perSecond(total, elapsed))
}

func (e *Engine) handleBench(args []string) bool {
	depth := depthArg(args, defaultBenchDepth)
	if len(args) > 1 {
		if n, err := strconv.Atoi(args[1]); err == nil && n >= 1 {
			e.searcher.SetThreads(n)
		}
	}

	// Progress lines would drown out the summary, and each position is searched
	// from a clean slate so that the total is reproducible.
	e.searcher.OnInfo = nil

	// Only the searches are timed. Each position starts from a cleared table so
	// that the node counts are reproducible, but emptying tens of megabytes is not
	// something an engine does between moves, so counting it here would flatter or
	// penalise changes that have nothing to do with search speed.
	var totalNodes uint64
	var elapsed time.Duration
	for i, fen := range benchPositions {
		position, err := engine.NewPositionFromFEN(fen)
		if err != nil {
			e.printf("info string bench: %v\n", err)
			continue
		}
		e.searcher.NewGame()

		start := time.Now()
		info := e.searcher.Search(position, search.Limits{Depth: depth})
		elapsed += time.Since(start)
		totalNodes += info.Nodes

		e.printf("position %d/%d: bestmove %s score %d nodes %d\n",
			i+1, len(benchPositions), info.BestMove(), info.Score, info.Nodes)
	}

	e.printf("\nbench: depth %d, threads %d, %d nodes, %v, %.0f nodes/s\n",
		depth, e.searcher.Threads(), totalNodes,
		elapsed.Round(time.Millisecond), perSecond(totalNodes, elapsed))
	return false
}

// benchPositions covers openings, middlegame tactics and endgames, so that the
// figure reflects the whole engine rather than one kind of position.
var benchPositions = []string{
	engine.StartFEN,
	"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
	"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
	"r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
	"4rrk1/pp1n3p/3q2pQ/2p1pb2/2PP4/2P3N1/P2B2PP/4RRK1 b - - 7 19",
	"r3r1k1/2p2ppp/p1p1bn2/8/1q2P3/2NPQN2/PPP3PP/R4RK1 b - - 2 15",
	"6k1/6p1/6Pp/ppp5/3pn2P/1P3K2/1PP2P2/3N4 b - - 0 1",
	"8/8/8/8/8/6k1/6p1/6K1 w - - 0 1",
	"8/2k5/4p3/1nB2p2/2K5/8/8/8 b - - 0 1",
}

// depthArg reads an optional depth argument.
func depthArg(args []string, fallback int) int {
	if len(args) == 0 {
		return fallback
	}
	n, err := strconv.Atoi(args[0])
	if err != nil || n < 1 {
		return fallback
	}
	return n
}

func perSecond(nodes uint64, elapsed time.Duration) float64 {
	if elapsed <= 0 {
		return 0
	}
	return float64(nodes) / elapsed.Seconds()
}
