// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

// Package engine implements the board representation, move generation and
// position handling for the chesssy chess engine.
//
// # Board representation
//
// Squares are indexed little-endian rank-file: A1 = 0, B1 = 1, ... H8 = 63.
// That layout makes "north" a left shift by 8, which keeps bitboard shifts cheap
// and readable. A [Position] holds one [Bitboard] per piece type and one per
// color, plus a piece-per-square array for fast lookups.
//
// # Reading order
//
//	piece.go, square.go   the value types: colors, pieces, squares, directions
//	bitboard.go           the square-set type and its operations
//	magic.go, attacks.go  precomputed attack tables, magic bitboards for sliders
//	rand.go               the fixed-seed generator those tables are built with
//	move.go, movelist.go  the packed 16-bit move and the generator's output buffer
//	position.go           the position type, its accessors and board mutation
//	fen.go                Forsyth-Edwards notation, in and out
//	castling.go           castling rights and the squares each one involves
//	zobrist.go            the keys that hash a position into 64 bits
//	makemove.go           playing and unplaying moves, incremental hashing
//	checks.go             attack, check and pin queries
//	movegen.go            legal move generation
//	draw.go               repetition, the fifty-move rule, insufficient material
//	see.go                static exchange evaluation
//	perft.go              the move generation correctness harness
//	errors.go             the sentinel errors parsing reports
//
// # Concurrency
//
// The lookup tables are immutable after package initialisation, so the package
// level functions are safe for concurrent use. A [Position] is mutable and is
// not: give each goroutine its own, for example with [Position.Clone].
package engine
