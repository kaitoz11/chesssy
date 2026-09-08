// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package uci

import (
	"strings"
	"testing"
)

func TestPositionCommands(t *testing.T) {
	tests := []struct {
		name     string
		commands []string
		wantFEN  string
	}{
		{
			name:     "startpos",
			commands: []string{"position startpos", "fen"},
			wantFEN:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		},
		{
			name:     "startpos with moves",
			commands: []string{"position startpos moves e2e4 e7e5 g1f3", "fen"},
			wantFEN:  "rnbqkbnr/pppp1ppp/8/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2",
		},
		{
			name:     "fen",
			commands: []string{"position fen 8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1", "fen"},
			wantFEN:  "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		},
		{
			name:     "fen with moves",
			commands: []string{"position fen 4k3/8/8/8/8/8/6P1/4K3 w - - 0 1 moves g2g4", "fen"},
			wantFEN:  "4k3/8/8/8/6P1/8/8/4K3 b - g3 0 1",
		},
		{
			name:     "promotion move",
			commands: []string{"position fen 4k3/6P1/8/8/8/8/8/4K3 w - - 0 1 moves g7g8q", "fen"},
			wantFEN:  "4k1Q1/8/8/8/8/8/8/4K3 b - - 0 1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := run(t, append(tc.commands, "quit")...)
			if !strings.Contains(got, tc.wantFEN) {
				t.Errorf("want FEN %q in output:\n%s", tc.wantFEN, got)
			}
		})
	}
}

func TestPositionErrorsAreReported(t *testing.T) {
	got := run(t, "position fen not-a-fen", "position startpos moves e2e5", "position", "quit")
	if strings.Count(got, "info string") < 3 {
		t.Errorf("expected an info string per bad command:\n%s", got)
	}
	// A rejected command must not crash the engine or lose the good position.
	if !strings.Contains(run(t, "position fen bad", "position startpos", "fen", "quit"), "rnbqkbnr") {
		t.Error("engine did not recover from an invalid position command")
	}
}
