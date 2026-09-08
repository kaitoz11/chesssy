// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runArgs drives the command line with no standard input and returns what it wrote.
func runArgs(t *testing.T, args ...string) string {
	t.Helper()
	var out bytes.Buffer
	if err := run(args, &out, strings.NewReader("")); err != nil {
		t.Fatalf("run(%v) = %v", args, err)
	}
	return out.String()
}

func TestSubcommands(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "no arguments speaks uci",
			args: nil,
			want: []string{"chesssy", "a UCI chess engine"},
		},
		{
			name: "uci",
			args: []string{"uci"},
			want: []string{"a UCI chess engine"},
		},
		{
			name: "version",
			args: []string{"version"},
			want: []string{"chesssy 1.0"},
		},
		{
			name: "perft counts the start position",
			args: []string{"perft", "-depth", "3"},
			want: []string{"a2a3: 380", "nodes: 8902", "depth 3"},
		},
		{
			name: "perft accepts a position",
			args: []string{"perft", "-depth", "3", "-fen", "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1"},
			want: []string{"nodes: 2812"},
		},
		{
			name: "bench reports a speed",
			args: []string{"bench", "-depth", "4"},
			want: []string{"bench: depth 4", "nodes/s"},
		},
		{
			name: "eval draws the board and scores it",
			args: []string{"eval", "-fen", "4k3/8/8/8/8/8/8/3QK3 w - - 0 1"},
			want: []string{"a b c d e f g h", "static eval"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := runArgs(t, tc.args...)
			for _, want := range tc.want {
				if !strings.Contains(got, want) {
					t.Errorf("output missing %q:\n%s", want, got)
				}
			}
		})
	}
}

func TestUnknownCommandFails(t *testing.T) {
	var out bytes.Buffer
	err := run([]string{"nonsense"}, &out, strings.NewReader(""))
	if err == nil {
		t.Fatal("expected an error for an unknown command")
	}
	if !strings.Contains(err.Error(), "usage:") {
		t.Errorf("error should include the usage line, got %v", err)
	}
}

func TestBadFlagFails(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"perft", "-nonsense"}, &out, strings.NewReader("")); err == nil {
		t.Fatal("expected an error for an unknown flag")
	}
}

func TestCPUProfileIsWritten(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cpu.prof")
	runArgs(t, "perft", "-depth", "4", "-cpuprofile", path)

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("no profile written: %v", err)
	}
	if info.Size() == 0 {
		t.Error("profile is empty")
	}
}

func TestCPUProfileToUnwritablePathFails(t *testing.T) {
	var out bytes.Buffer
	err := run([]string{"perft", "-cpuprofile", filepath.Join(t.TempDir(), "missing", "cpu.prof")},
		&out, strings.NewReader(""))
	if err == nil {
		t.Fatal("expected an error for an unwritable profile path")
	}
}

func TestUCIReadsStandardInput(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("position startpos\ngo depth 4\nquit\n")
	if err := run(nil, &out, in); err != nil {
		t.Fatalf("run: %v", err)
	}
	if got := out.String(); !strings.Contains(got, "bestmove") {
		t.Errorf("expected a bestmove in:\n%s", got)
	}
}
