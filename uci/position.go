// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package uci

import (
	"fmt"
	"slices"
	"strings"

	"github.com/kaitoz11/chesssy/engine"
)

// The position command sets the board, either from the standard start or from a
// FEN, then optionally replays a list of moves:
//
//	position startpos moves e2e4 e7e5
//	position fen 8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1
//
// Replaying the moves rather than sending the resulting FEN is what lets the engine
// see repetitions, so the move list is applied to the position it builds.

func (e *Engine) handlePosition(args []string) bool {
	e.stopSearch()
	if err := e.setPosition(args); err != nil {
		// A bad position command is reported and ignored: the previous position
		// stays in place, so the session can continue.
		e.printf("info string %v\n", err)
	}
	return false
}

func (e *Engine) setPosition(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("position: expected startpos or fen")
	}

	// Everything before the moves keyword describes the starting board.
	board, moves := args, []string(nil)
	if i := slices.Index(args, "moves"); i >= 0 {
		board, moves = args[:i], args[i+1:]
	}

	position, err := parseBoard(board)
	if err != nil {
		return err
	}
	for _, token := range moves {
		m, err := position.ParseMove(token)
		if err != nil {
			return fmt.Errorf("position: %w", err)
		}
		position.MakeMove(m)
	}

	e.position = position
	return nil
}

// parseBoard builds the starting board of a position command.
func parseBoard(args []string) (*engine.Position, error) {
	switch args[0] {
	case "startpos":
		return engine.NewPosition(), nil
	case "fen":
		return engine.NewPositionFromFEN(strings.Join(args[1:], " "))
	default:
		return nil, fmt.Errorf("position: expected startpos or fen, got %q", args[0])
	}
}
