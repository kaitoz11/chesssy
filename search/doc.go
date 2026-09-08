// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

// Package search picks moves. It evaluates positions statically and searches the
// move tree with a fail-soft principal variation search, an alpha-beta variant,
// driven by iterative deepening.
//
// # Using it
//
//	s := search.NewSearcher(search.DefaultHashSizeMB)
//	s.OnInfo = func(info search.Info) { log.Println(info.Depth, info.Score) }
//	info := s.Search(position, search.Limits{Depth: 12})
//	best := info.BestMove()
//
// Reuse one [Searcher] for a whole game: the transposition table and the move
// ordering heuristics it accumulates are what make each move cheaper than the
// last. One Searcher runs one search at a time.
//
// # Reading order
//
//	score.go       what the numbers mean, including mate scores
//	params.go      every tunable constant, in one place
//	eval.go        static evaluation, blended between middlegame and endgame
//	psqt.go        the piece-square tables eval.go reads
//	pawns.go       pawn structure, memoised in a cache
//	tt.go          the transposition table
//	movepick.go    the order moves are tried in
//	history.go     the heuristics that inform that order
//	worker.go      one search thread and everything private to it
//	searcher.go    the public API, and the thread pool behind it
//	search.go      iterative deepening and aspiration windows
//	negamax.go     the main search loop
//	pruning.go     the bets that let it skip and shorten work
//	quiescence.go  the capture-only search at the leaves
//	limits.go      when to stop
package search
